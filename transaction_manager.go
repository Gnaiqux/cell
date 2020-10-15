package main

import (
	"crypto/tls"
	"fmt"
	"github.com/project-nano/cell/service"
	"github.com/project-nano/cell/task"
	"github.com/project-nano/framework"
	"math/rand"
	"net/http"
	"time"
)

type TransactionManager struct {
	*framework.TransactionEngine
}

func CreateTransactionManager(sender framework.MessageSender, instanceModule *service.InstanceManager,
	storageModule *service.StorageManager, networkModule *service.NetworkManager) (manager *TransactionManager, err error) {
	var engine *framework.TransactionEngine
	if engine, err = framework.CreateTransactionEngine(); err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	generator := rand.New(rand.NewSource(time.Now().UnixNano()))

	manager = &TransactionManager{engine}
	if err = manager.RegisterExecutor(framework.GetComputePoolCellRequest,
		&task.GetCellInfoExecutor{sender, instanceModule, storageModule, networkModule}); err != nil{
		return nil, err
	}

	if err = manager.RegisterExecutor(framework.CreateGuestRequest,
		&task.CreateInstanceExecutor{sender, instanceModule, storageModule, networkModule, generator}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.DeleteGuestRequest,
		&task.DeleteInstanceExecutor{sender, instanceModule, storageModule, networkModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.GetGuestRequest,
		&task.GetInstanceConfigExecutor{sender, instanceModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.GetInstanceStatusRequest,
		&task.GetInstanceStatusExecutor{sender, instanceModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.StartInstanceRequest,
		&task.StartInstanceExecutor{sender, instanceModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.StopInstanceRequest,
		&task.StopInstanceExecutor{sender, instanceModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.AttachInstanceRequest,
		&task.AttachInstanceExecutor{sender, instanceModule, storageModule, networkModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.DetachInstanceRequest,
		&task.DetachInstanceExecutor{sender, instanceModule, storageModule, networkModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.ModifyGuestNameRequest,
		&task.ModifyGuestNameExecutor{sender, instanceModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.ModifyCoreRequest,
		&task.ModifyGuestCoreExecutor{sender, instanceModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.ModifyMemoryRequest,
		&task.ModifyGuestMemoryExecutor{sender, instanceModule}); err != nil{
		return nil, err
	}

	if err = manager.RegisterExecutor(framework.ModifyPriorityRequest,
		&task.ModifyCPUPriorityExecutor{sender, instanceModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.ModifyDiskThresholdRequest,
		&task.ModifyDiskThresholdExecutor{sender, instanceModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.ModifyNetworkThresholdRequest,
		&task.ModifyNetworkThresholdExecutor{sender, instanceModule}); err != nil{
		return nil, err
	}

	if err = manager.RegisterExecutor(framework.ModifyAuthRequest,
		&task.ModifyGuestPasswordExecutor{sender, instanceModule, generator}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.GetAuthRequest,
		&task.GetGuestPasswordExecutor{sender, instanceModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.ResetSystemRequest,
		&task.ResetGuestSystemExecutor{sender, instanceModule, storageModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.InsertMediaRequest,
		&task.InsertMediaCoreExecutor{sender, instanceModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.EjectMediaRequest,
		&task.EjectMediaCoreExecutor{sender, instanceModule}); err != nil{
		return nil, err
	}

	if err = manager.RegisterExecutor(framework.ComputePoolReadyEvent,
		&task.HandleComputePoolReadyExecutor{sender, instanceModule, storageModule, networkModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.ComputeCellRemovedEvent,
		&task.HandleComputeCellRemovedExecutor{sender, instanceModule, storageModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.CreateDiskImageRequest,
		&task.CreateDiskImageExecutor{sender, instanceModule, storageModule, client}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.ResizeDiskRequest,
		&task.ResizeGuestVolumeExecutor{sender, instanceModule, storageModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.ShrinkDiskRequest,
		&task.ShrinkGuestVolumeExecutor{sender, instanceModule, storageModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.QuerySnapshotRequest,
		&task.QuerySnapshotExecutor{sender, storageModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.GetSnapshotRequest,
		&task.GetSnapshotExecutor{sender, storageModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.AddressPoolChangedEvent,
		&task.HandleAddressPoolChangedExecutor{instanceModule, networkModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.CreateSnapshotRequest,
		&task.CreateSnapshotExecutor{sender, instanceModule, storageModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.DeleteSnapshotRequest,
		&task.DeleteSnapshotExecutor{sender, instanceModule, storageModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.RestoreSnapshotRequest,
		&task.RestoreSnapshotExecutor{sender, instanceModule, storageModule}); err != nil{
		return nil, err
	}
	if err = manager.RegisterExecutor(framework.ResetSecretRequest,
		&task.ResetMonitorSecretExecutor{
			Sender:         sender,
			InstanceModule: instanceModule,
		}); err != nil{
		err = fmt.Errorf("register reset monitor secret fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(framework.QueryCellStorageRequest,
		&task.QueryStoragePathExecutor{
			Sender:  sender,
			Storage: storageModule,
		}); err != nil{
		err = fmt.Errorf("register query storage paths fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(framework.ModifyCellStorageRequest,
		&task.ChangeStoragePathExecutor{
			Sender:  sender,
			Storage: storageModule,
		}); err != nil{
		err = fmt.Errorf("register change storage path fail: %s", err.Error())
		return
	}
	return manager, nil
}
