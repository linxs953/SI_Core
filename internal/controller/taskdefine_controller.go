/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	lctv1 "kube.inspect/api/v1"
)

const (
	taskDefineFinalizer = "taskdefine.lct.kube.inspect/finalizer"
)

// TaskDefineReconciler reconciles a TaskDefine object
type TaskDefineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=lct.kube.inspect,resources=taskdefines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=lct.kube.inspect,resources=taskdefines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=lct.kube.inspect,resources=taskdefines/finalizers,verbs=update

func (r *TaskDefineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling TaskDefine", "namespace", req.Namespace, "name", req.Name)

	// 获取 TaskDefine 实例
	taskDefine := &lctv1.TaskDefine{}
	if err := r.Get(ctx, req.NamespacedName, taskDefine); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("TaskDefine resource not found, ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get TaskDefine")
		return ctrl.Result{RequeueAfter: time.Second * 5}, err
	}

	// 处理删除操作
	if !taskDefine.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, taskDefine)
	}

	// 处理创建/更新操作
	return r.handleCreateOrUpdate(ctx, taskDefine)
}

// handleDeletion 处理资源删除
func (r *TaskDefineReconciler) handleDeletion(ctx context.Context, taskDefine *lctv1.TaskDefine) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling deletion", "name", taskDefine.Name)

	// 检查 finalizer
	if controllerutil.ContainsFinalizer(taskDefine, taskDefineFinalizer) {
		// 执行清理操作
		if err := r.cleanupResources(ctx, taskDefine); err != nil {
			logger.Error(err, "Failed to cleanup resources")
			return ctrl.Result{RequeueAfter: time.Second * 5}, err
		}

		// 移除 finalizer
		controllerutil.RemoveFinalizer(taskDefine, taskDefineFinalizer)
		if err := r.Update(ctx, taskDefine); err != nil {
			if errors.IsConflict(err) {
				logger.Info("Resource version conflict when removing finalizer, will retry")
				return ctrl.Result{RequeueAfter: time.Second * 2}, nil
			}
			logger.Error(err, "Failed to remove finalizer")
			return ctrl.Result{RequeueAfter: time.Second * 5}, err
		}
		logger.Info("Successfully removed finalizer")
	}

	return ctrl.Result{}, nil
}

// handleCreateOrUpdate 处理资源创建和更新
func (r *TaskDefineReconciler) handleCreateOrUpdate(ctx context.Context, taskDefine *lctv1.TaskDefine) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling create/update", "name", taskDefine.Name)

	// 添加 finalizer（如果不存在）
	if !controllerutil.ContainsFinalizer(taskDefine, taskDefineFinalizer) {
		controllerutil.AddFinalizer(taskDefine, taskDefineFinalizer)
		if err := r.Update(ctx, taskDefine); err != nil {
			if errors.IsConflict(err) {
				logger.Info("Resource version conflict when adding finalizer, will retry")
				return ctrl.Result{RequeueAfter: time.Second * 2}, nil
			}
			logger.Error(err, "Failed to add finalizer")
			return ctrl.Result{RequeueAfter: time.Second * 5}, err
		}
		logger.Info("Successfully added finalizer")
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	// 验证和更新状态
	if err := r.validateAndUpdateStatus(ctx, taskDefine); err != nil {
		logger.Error(err, "Failed to validate and update status")
		return ctrl.Result{RequeueAfter: time.Second * 5}, err
	}

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

// cleanupResources 清理相关资源
func (r *TaskDefineReconciler) cleanupResources(ctx context.Context, taskDefine *lctv1.TaskDefine) error {
	logger := log.FromContext(ctx)
	logger.Info("Cleaning up resources", "name", taskDefine.Name)

	// TODO: 添加资源清理逻辑
	return nil
}

// validateAndUpdateStatus 验证资源并更新状态
func (r *TaskDefineReconciler) validateAndUpdateStatus(ctx context.Context, taskDefine *lctv1.TaskDefine) error {
	logger := log.FromContext(ctx)
	statusChanged := false

	// 检查状态是否需要初始化
	if taskDefine.Status.State == "" {
		taskDefine.Status.State = "Pending"
		taskDefine.Status.Message = "Initializing"
		statusChanged = true
	}

	// 验证规范字段
	if err := r.validateSpec(taskDefine); err != nil {
		if taskDefine.Status.State != "Invalid" || taskDefine.Status.Message != err.Error() {
			taskDefine.Status.State = "Invalid"
			taskDefine.Status.Message = fmt.Sprintf("Validation failed: %v", err)
			statusChanged = true
		}
	} else {
		// 验证成功，更新状态为 Ready
		if taskDefine.Status.State != "Ready" {
			taskDefine.Status.State = "Ready"
			taskDefine.Status.Message = "Resource is ready"
			statusChanged = true
		}
	}

	// 只在状态发生变化时更新
	if statusChanged {
		taskDefine.Status.LastUpdated = time.Now().Format(time.RFC3339)
		if err := r.Status().Update(ctx, taskDefine); err != nil {
			if errors.IsConflict(err) {
				logger.Info("Resource version conflict when updating status, will retry")
				return err
			}
			logger.Error(err, "Failed to update status")
			return err
		}
		logger.Info("Successfully updated status",
			"name", taskDefine.Name,
			"state", taskDefine.Status.State,
			"message", taskDefine.Status.Message)
	}

	return nil
}

// validateSpec 验证规范字段
func (r *TaskDefineReconciler) validateSpec(taskDefine *lctv1.TaskDefine) error {
	// 验证 IDL 类型
	if taskDefine.Spec.IdlType != "" {
		switch taskDefine.Spec.IdlType {
		case "1", "2", "3":
			// 有效的 IDL 类型
		default:
			return fmt.Errorf("invalid IdlType: %s, must be one of: 1, 2, 3", taskDefine.Spec.IdlType)
		}
	}

	// 验证 RelatedImage
	if taskDefine.Spec.RelatedImage.Builder == "" {
		return fmt.Errorf("relatedImage.builder is required")
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TaskDefineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&lctv1.TaskDefine{}).
		Complete(r)
}
