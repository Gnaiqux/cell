package task

import (
	"errors"
	"fmt"
	"github.com/project-nano/cell/service"
	"github.com/project-nano/framework"
	"log"
	"math/rand"
	"time"
)

type CreateInstanceExecutor struct {
	Sender          framework.MessageSender
	InstanceModule  service.InstanceModule
	StorageModule   service.StorageModule
	NetworkModule   service.NetworkModule
	RandomGenerator *rand.Rand
}

func (executor *CreateInstanceExecutor) Execute(id framework.SessionID, request framework.Message,
	incoming chan framework.Message, terminate chan bool) (err error) {
	resp, _ := framework.CreateJsonMessage(framework.CreateGuestResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	var config = service.GuestConfig{}
	config.Initialized = false
	//full name: group.instanceName
	if config.Name, err = request.GetString(framework.ParamKeyName); err != nil {
		err = fmt.Errorf("get name fail: %s", err.Error())
		return executor.ResponseFail(resp, err.Error(), request.GetSender())
	}
	if config.ID, err = request.GetString(framework.ParamKeyInstance); err != nil {
		err = fmt.Errorf("get id fail: %s", err.Error())
		return executor.ResponseFail(resp, err.Error(), request.GetSender())
	}
	if config.User, err = request.GetString(framework.ParamKeyUser); err != nil {
		err = fmt.Errorf("get user fail: %s", err.Error())
		return executor.ResponseFail(resp, err.Error(), request.GetSender())
	}
	if config.Group, err = request.GetString(framework.ParamKeyGroup); err != nil {
		err = fmt.Errorf("get group fail: %s", err.Error())
		return executor.ResponseFail(resp, err.Error(), request.GetSender())
	}

	if config.Cores, err = request.GetUInt(framework.ParamKeyCore); err != nil {
		err = fmt.Errorf("get cores fail: %s", err.Error())
		return executor.ResponseFail(resp, err.Error(), request.GetSender())
	}

	if config.Memory, err = request.GetUInt(framework.ParamKeyMemory); err != nil {
		err = fmt.Errorf("get memory fail: %s", err.Error())
		return executor.ResponseFail(resp, err.Error(), request.GetSender())
	}
	var diskSize []uint64
	if diskSize, err = request.GetUIntArray(framework.ParamKeyDisk); err != nil {
		err = fmt.Errorf("get disk size fail: %s", err.Error())
		return executor.ResponseFail(resp, err.Error(), request.GetSender())
	}
	if config.AutoStart, err = request.GetBoolean(framework.ParamKeyOption); err != nil {
		err = fmt.Errorf("get auto start flag fail: %s", err.Error())
		return executor.ResponseFail(resp, err.Error(), request.GetSender())
	}
	if config.AuthUser, err = request.GetString(framework.ParamKeyAdmin); err != nil {
		err = fmt.Errorf("get admin name fail: %s", err.Error())
		return executor.ResponseFail(resp, err.Error(), request.GetSender())
	}
	var templateOptions []uint64
	if templateOptions, err = request.GetUIntArray(framework.ParamKeyTemplate); err != nil {
		err = fmt.Errorf("get template fail: %s", err.Error())
		return executor.ResponseFail(resp, err.Error(), request.GetSender())
	} else {
		const (
			OptionOffsetOS = iota
			OptionOffsetDisk
			OptionOffsetNetwork
			OptionOffsetDisplay
			OptionOffsetControl
			OptionOffsetUSB
			OptionOffsetTablet
			ValidOptionCount
		)
		if ValidOptionCount != len(templateOptions) {
			err = fmt.Errorf("template options count mismatch %d / %d", len(templateOptions), ValidOptionCount)
			return executor.ResponseFail(resp, err.Error(), request.GetSender())
		}
		var t = service.HardwareTemplate{
			OperatingSystem: service.TemplateOperatingSystem(templateOptions[OptionOffsetOS]).ToString(),
			Disk:            service.TemplateDiskDriver(templateOptions[OptionOffsetDisk]).ToString(),
			Network:         service.TemplateNetworkModel(templateOptions[OptionOffsetNetwork]).ToString(),
			Display:         service.TemplateDisplayDriver(templateOptions[OptionOffsetDisplay]).ToString(),
			Control:         service.TemplateRemoteControl(templateOptions[OptionOffsetControl]).ToString(),
			USB:             service.TemplateUSBModel(templateOptions[OptionOffsetUSB]).ToString(),
			Tablet:          service.TemplateTabletModel(templateOptions[OptionOffsetTablet]).ToString(),
		}
		config.Template = &t
	}

	if modeArray, err := request.GetUIntArray(framework.ParamKeyMode); err != nil {
		return err
	} else {
		const (
			ValidModeCount = 2 //[network, storage]
		)
		if ValidModeCount != len(modeArray) {
			return fmt.Errorf("unexpect mode params count %d", len(modeArray))
		}
		config.NetworkMode = service.InstanceNetworkMode(modeArray[0])
		config.StorageMode = service.InstanceStorageMode(modeArray[1])
	}
	var cloneFromImage = false
	var imageID, mediaHost string
	var mediaPort, imageSize uint
	if imageID, err = request.GetString(framework.ParamKeyImage); err == nil {
		cloneFromImage = true
		if mediaHost, err = request.GetString(framework.ParamKeyHost); err != nil {
			err = fmt.Errorf("get media host fail: %s", err.Error())
			return executor.ResponseFail(resp, err.Error(), request.GetSender())
		}
		if mediaPort, err = request.GetUInt(framework.ParamKeyPort); err != nil {
			err = fmt.Errorf("get media port fail: %s", err.Error())
			return executor.ResponseFail(resp, err.Error(), request.GetSender())
		}
		if imageSize, err = request.GetUInt(framework.ParamKeySize); err != nil {
			err = fmt.Errorf("get image size fail: %s", err.Error())
			return executor.ResponseFail(resp, err.Error(), request.GetSender())
		}
	}
	if assignedAddress, err := request.GetStringArray(framework.ParamKeyAddress); err == nil {
		const (
			ValidAssignedLength = 2
		)
		if len(assignedAddress) != ValidAssignedLength {
			err = fmt.Errorf("unexpect assigned addresses count %d", len(assignedAddress))
			return executor.ResponseFail(resp, err.Error(), request.GetSender())
		}
		config.InternalAddress = assignedAddress[0]
		config.ExternalAddress = assignedAddress[1]
	}
	//QoS
	{
		priorityValue, _ := request.GetUInt(framework.ParamKeyPriority)
		config.CPUPriority = service.PriorityEnum(priorityValue)
		if limitParameters, err := request.GetUIntArray(framework.ParamKeyLimit); err == nil {
			const (
				ReadSpeedOffset = iota
				WriteSpeedOffset
				ReadIOPSOffset
				WriteIOPSOffset
				ReceiveOffset
				SendOffset
				ValidLimitParametersCount = 6
			)

			if ValidLimitParametersCount != len(limitParameters) {
				err = fmt.Errorf("invalid QoS parameters count %d", len(limitParameters))
				return executor.ResponseFail(resp, err.Error(), request.GetSender())
			}
			config.ReadSpeed = limitParameters[ReadSpeedOffset]
			config.WriteSpeed = limitParameters[WriteSpeedOffset]
			config.ReadIOPS = limitParameters[ReadIOPSOffset]
			config.WriteIOPS = limitParameters[WriteIOPSOffset]
			config.ReceiveSpeed = limitParameters[ReceiveOffset]
			config.SendSpeed = limitParameters[SendOffset]
		}
	}
	log.Printf("[%08X] request create instance '%s' ( id: %s ) from %s.[%08X]", id,
		config.Name, config.ID, request.GetSender(), request.GetFromSession())

	log.Printf("[%08X] require %d cores, %d MB memory", id, config.Cores, config.Memory>>20)
	log.Printf("[%08X] IO limit: read %d, write %d per second, network limit: recv %d Kps, send %d Kps",
		id, config.ReadIOPS, config.WriteIOPS, config.ReceiveSpeed>>10, config.SendSpeed>>10)

	//Security Policy
	{
		const (
			offsetAccept = iota
			offsetProtocol
			offsetFrom
			offsetTo
			offsetPort
			validPolicyElementCount = 5 //accept,protocol,from,to,port
		)
		var policyParameters []uint64
		if policyParameters, err = request.GetUIntArray(framework.ParamKeyPolicy); nil == err {
			if 0 != len(policyParameters)%validPolicyElementCount {
				err = fmt.Errorf("invalid policy parameters count %d", len(policyParameters))
				return executor.ResponseFail(resp, err.Error(), request.GetSender())
			}
			var parameterCount = len(policyParameters)
			var ruleCount = parameterCount / validPolicyElementCount
			var securityPolicy service.SecurityPolicy
			if securityPolicy.Accept, err = request.GetBoolean(framework.ParamKeyAction); err != nil {
				err = fmt.Errorf("get default security action fail: %s", err.Error())
				return executor.ResponseFail(resp, err.Error(), request.GetSender())
			}
			for start := 0; start < parameterCount; start += validPolicyElementCount {
				var rule service.SecurityPolicyRule
				if service.PolicyRuleActionAccept == policyParameters[start+offsetAccept] {
					rule.Accept = true
				} else {
					rule.Accept = false
				}
				switch policyParameters[start+offsetProtocol] {
				case service.PolicyRuleProtocolIndexTCP:
					rule.Protocol = service.PolicyRuleProtocolTCP
				case service.PolicyRuleProtocolIndexUDP:
					rule.Protocol = service.PolicyRuleProtocolUDP
				case service.PolicyRuleProtocolIndexICMP:
					rule.Protocol = service.PolicyRuleProtocolICMP
				default:
					err = fmt.Errorf("invalid security protocol index %d", policyParameters[start+offsetProtocol])
					return executor.ResponseFail(resp, err.Error(), request.GetSender())
				}
				rule.SourceAddress = service.UInt32ToIPv4(uint32(policyParameters[start+offsetFrom]))
				rule.TargetAddress = service.UInt32ToIPv4(uint32(policyParameters[start+offsetTo]))
				rule.TargetPort = uint(policyParameters[start+offsetPort])
				securityPolicy.Rules = append(securityPolicy.Rules, rule)
				//log.Printf("[%08X] debug: policy parameters %d, %d, %d, %d, %d",
				//	id, policyParameters[start + offsetAccept], policyParameters[start + offsetProtocol], policyParameters[start + offsetFrom],
				//	policyParameters[start + offsetTo], policyParameters[start + offsetPort])
			}
			config.Security = &securityPolicy
			if ruleCount > 0 {
				log.Printf("[%08X] %d security rule(s) with default accept: %t",
					id, len(securityPolicy.Rules), securityPolicy.Accept)
			}
		}
	}

	diskCount := len(diskSize)
	if 0 == diskCount {
		err = errors.New("must specify disk size")
		return executor.ResponseFail(resp, err.Error(), request.GetSender())
	}
	var systemSize = diskSize[0]
	log.Printf("[%08X] system disk %d GB", id, systemSize>>30)

	var dataSize []uint64
	if len(diskSize) > 1 {
		dataSize = diskSize[1:]
		index := 0
		for _, volSize := range dataSize {
			log.Printf("[%08X] data disk %d: %d GB", id, index, volSize>>30)
			index++
		}
	}

	log.Printf("[%08X] network mode %d, storage mode %d, auto start : %t", id,
		config.NetworkMode, config.StorageMode, config.AutoStart)

	log.Printf("[%08X] operating system type: %s, admin name '%s'", id, config.Template.OperatingSystem, config.AuthUser)

	{
		//check modules & ci params
		const (
			ModuleQEMU      = "qemu"
			ModuleCloudInit = "cloud-init"
		)
		var ValidModules = map[string]bool{ModuleQEMU: true, ModuleCloudInit: true}
		modules, _ := request.GetStringArray(framework.ParamKeyModule)
		for _, moduleName := range modules {
			if _, exists := ValidModules[moduleName]; !exists {
				err = fmt.Errorf("invalid module '%s'", moduleName)
				log.Printf("[%08X] verify modules fail: %s", id, err.Error())
				return executor.ResponseFail(resp, err.Error(), request.GetSender())
			} else if ModuleQEMU == moduleName {
				config.QEMUAvailable = true
				log.Printf("[%08X] qemu module available", id)
			} else if ModuleCloudInit == moduleName {
				config.CloudInitAvailable = true
				log.Printf("[%08X] cloud-init module available", id)
			}
		}
		{

			//flags
			const (
				LoginEnableFlag = 0
				ValidFlagLength = 1
			)

			const (
				RootLoginDisabled = iota
				RootLoginEnabled
			)
			flags, err := request.GetUIntArray(framework.ParamKeyFlag)
			if err != nil {
				err = fmt.Errorf("get flags fail: %s", err.Error())
				return executor.ResponseFail(resp, err.Error(), request.GetSender())
			}
			if ValidFlagLength != len(flags) {
				err = fmt.Errorf("invalid flags count %d", len(flags))
				log.Printf("[%08X] verify flags fail: %s", id, err.Error())
				return executor.ResponseFail(resp, err.Error(), request.GetSender())
			}
			if RootLoginEnabled == flags[LoginEnableFlag] {
				config.RootLoginEnabled = true
				log.Printf("[%08X] remote root access via ssh enabled", id)
			} else {
				config.RootLoginEnabled = false
				log.Printf("[%08X] remote root access via ssh disabled", id)
			}
		}
		//ci params
		if config.CloudInitAvailable {
			const (
				PasswordLength  = 10
				DefaultDataPath = "/opt/data"
			)

			password, err := request.GetString(framework.ParamKeySecret)
			if err != nil {
				err = fmt.Errorf("get user password fail: %s", err.Error())
				return executor.ResponseFail(resp, err.Error(), request.GetSender())
			}

			if 0 == len(password) {
				password = executor.generatePassword(PasswordLength)
				log.Printf("[%08X] %d byte(s) password generated", id, len(password))
			}
			config.AuthSecret = password
			config.DataPath, err = request.GetString(framework.ParamKeyPath)
			if err != nil {
				err = fmt.Errorf("get data path fail: %s", err.Error())
				return executor.ResponseFail(resp, err.Error(), request.GetSender())
			}
			if "" == config.DataPath {
				config.DataPath = DefaultDataPath
			}
			log.Printf("[%08X] data disk mount path '%s'", id, config.DataPath)
		}
	}

	{
		//network prepare
		if "" == config.HardwareAddress {
			mac, err := executor.generateMacAddress()
			if err != nil {
				err = fmt.Errorf("get MAC address fail: %s", err.Error())
				return executor.ResponseFail(resp, err.Error(), request.GetSender())
			}
			config.HardwareAddress = mac
			log.Printf("[%08X] mac '%s' generated", id, mac)
		}

		switch config.NetworkMode {
		case service.NetworkModePlain:
			{
				//find bridge
				var respChan = make(chan service.NetworkResult)
				executor.NetworkModule.GetCurrentConfig(respChan)
				result := <-respChan
				if result.Error != nil {
					err = result.Error
					log.Printf("[%08X] get default bridge fail: %s", id, err.Error())
					return executor.ResponseFail(resp, err.Error(), request.GetSender())
				}
				config.NetworkSource = result.Name
				config.AddressAllocation = result.Allocation
				log.Printf("[%08X] network bridge '%s' (mode '%s') allocated for instance '%s'",
					id, config.NetworkSource, config.AddressAllocation, config.Name)
			}
			{
				//monitor port
				var respChan = make(chan service.NetworkResult)
				executor.NetworkModule.AllocateInstanceResource(config.ID, config.HardwareAddress, config.InternalAddress, config.ExternalAddress, respChan)
				result := <-respChan
				if result.Error != nil {
					err = result.Error
					log.Printf("[%08X] allocate monitor port fail: %s", id, err.Error())
					return executor.ResponseFail(resp, err.Error(), request.GetSender())
				}
				config.MonitorPort = uint(result.MonitorPort)
				log.Printf("[%08X] monitor port %d allocated", id, config.MonitorPort)
			}

			break
		default:
			err = fmt.Errorf("unsupported network mode %d", config.NetworkMode)
			return executor.ResponseFail(resp, err.Error(), request.GetSender())
		}

	}

	var volGroup = config.ID
	{
		//create storage volumes
		respChan := make(chan service.StorageResult)
		var bootType = service.BootTypeNone
		if config.CloudInitAvailable {
			bootType = service.BootTypeCloudInit
		}

		executor.StorageModule.CreateVolumes(volGroup, systemSize, dataSize, bootType, respChan)

		result := <-respChan
		if result.Error != nil {
			err = result.Error
			log.Printf("[%08X] create volumes fail: %s", id, err.Error())
			executor.ReleaseResource(id, config.ID, true, false, false)
			return executor.ResponseFail(resp, err.Error(), request.GetSender())
		}
		config.StoragePool = result.Pool
		config.StorageVolumes = result.Volumes
		config.Disks = diskSize
		if config.CloudInitAvailable {
			config.BootImage = result.Image
		}
		log.Printf("[%08X] %d volumes allocated in pool '%s' with group '%s'", id, len(config.StorageVolumes), config.StoragePool, volGroup)
	}
	{
		var errChan = make(chan error, 1)
		const (
			MonitorSecretLength = 8
		)
		config.MonitorSecret = executor.generatePassword(MonitorSecretLength)
		executor.InstanceModule.CreateInstance(config, errChan)
		err = <-errChan
		if err != nil {
			executor.ReleaseResource(id, config.ID, true, true, false)
			log.Printf("[%08X] create instance fail: %s", id, err.Error())
			return executor.ResponseFail(resp, err.Error(), request.GetSender())
		}
		//send create response
		resp.SetString(framework.ParamKeyInstance, config.ID)
		resp.SetBoolean(framework.ParamKeyEnable, config.Created) //created
		resp.SetSuccess(true)
		if err = executor.Sender.SendMessage(resp, request.GetSender()); err != nil {
			log.Printf("[%08X] warning: send response fail: %s", id, err.Error())
			return err
		}
	}

	if !cloneFromImage {
		//finished
		event, _ := framework.CreateJsonMessage(framework.GuestCreatedEvent)
		event.SetFromSession(id)
		event.SetString(framework.ParamKeyInstance, config.ID)
		event.SetUInt(framework.ParamKeyMonitor, config.MonitorPort)
		event.SetString(framework.ParamKeySecret, config.MonitorSecret)
		event.SetString(framework.ParamKeyHardware, config.HardwareAddress)
		if err = executor.Sender.SendMessage(event, request.GetSender()); err != nil {
			log.Printf("[%08X] warning: notify instance created fail: %s", id, err.Error())
		}
		if config.AutoStart {
			executor.startAutoStartInstance(id, config.ID, config.Name, request.GetSender())
		}
		return nil
	}
	//begin clone image
	{
		event, _ := framework.CreateJsonMessage(framework.GuestUpdatedEvent)
		event.SetSuccess(true)
		event.SetFromSession(id)
		event.SetString(framework.ParamKeyInstance, config.ID)

		var targetVol = config.StorageVolumes[0]
		var startChan = make(chan error, 1)
		var progressChan = make(chan uint, 1)
		var resultChan = make(chan service.StorageResult, 1)
		executor.StorageModule.ReadDiskImage(id, config.ID, targetVol, imageID, uint64(systemSize), uint64(imageSize), mediaHost, mediaPort,
			startChan, progressChan, resultChan)
		{
			var timer = time.NewTimer(service.GetConfigurator().GetOperateTimeout())
			select {
			case err = <-startChan:
				if err != nil {
					log.Printf("[%08X] start disk image cloning fail: %s", id, err.Error())
					executor.ReleaseResource(id, config.ID, true, true, true)
					return executor.ResponseFail(event, err.Error(), request.GetSender())
				}
				log.Printf("[%08X] disk image cloning started", id)

			case <-timer.C:
				//wait start timeout
				err = errors.New("start clone disk image timeout")
				executor.ReleaseResource(id, config.ID, true, true, true)
				return executor.ResponseFail(event, err.Error(), request.GetSender())
			}
		}

		const (
			CheckInterval = 2 * time.Second
		)

		var latestUpdate = time.Now()
		var ticker = time.NewTicker(CheckInterval)
		for {
			select {
			case <-ticker.C:
				//check
				if time.Now().After(latestUpdate.Add(service.GetConfigurator().GetOperateTimeout())) {
					//timeout
					err = errors.New("timeout")
					log.Printf("[%08X] clone disk image fail: %s", id, err.Error())
					executor.ReleaseResource(id, config.ID, true, true, true)
					return executor.ResponseFail(event, err.Error(), request.GetSender())
				}
			case progress := <-progressChan:
				latestUpdate = time.Now()
				event.SetUInt(framework.ParamKeyProgress, progress)
				log.Printf("[%08X] progress => %d %%", id, progress)
				if err = executor.Sender.SendMessage(event, request.GetSender()); err != nil {
					log.Printf("[%08X] warning: notify progress fail: %s", id, err.Error())
				}
			case result := <-resultChan:
				err = result.Error
				if err != nil {
					log.Printf("[%08X] clone disk image fail: %s", id, err.Error())
					executor.ReleaseResource(id, config.ID, true, true, true)
					return executor.ResponseFail(event, err.Error(), request.GetSender())
				}
				log.Printf("[%08X] clone disk image success, %d MB in size", id, result.Size>>20)
				//notify guest created
				created, _ := framework.CreateJsonMessage(framework.GuestCreatedEvent)
				created.SetSuccess(true)
				created.SetFromSession(id)
				created.SetString(framework.ParamKeyInstance, config.ID)
				created.SetUInt(framework.ParamKeyMonitor, config.MonitorPort)
				created.SetString(framework.ParamKeySecret, config.MonitorSecret)
				created.SetString(framework.ParamKeyHardware, config.HardwareAddress)

				if err = executor.Sender.SendMessage(created, request.GetSender()); err != nil {
					log.Printf("[%08X] warning: notify instance created fail: %s", id, err.Error())
				}
				if config.AutoStart {
					executor.startAutoStartInstance(id, config.ID, config.Name, request.GetSender())
				}
				return nil
			}
		}
	}

}

func (executor *CreateInstanceExecutor) ReleaseResource(id framework.SessionID, guestID string, clearNetwork, clearVolumes, clearInstance bool) {
	if clearInstance {
		executor.ReleaseInstance(id, guestID)
	}
	if clearVolumes {
		executor.ReleaseVolumes(id, guestID)
	}
	if clearNetwork {
		executor.ReleaseNetworkResource(id, guestID)
	}
}

func (executor *CreateInstanceExecutor) ReleaseInstance(id framework.SessionID, instance string) {
	resp := make(chan error)
	executor.InstanceModule.DeleteInstance(instance, resp)
	err := <-resp
	if err != nil {
		log.Printf("[%08X] warning: release instance fail: %s", id, err.Error())
	}
}

func (executor *CreateInstanceExecutor) ReleaseVolumes(id framework.SessionID, groupName string) {
	resp := make(chan error)
	executor.StorageModule.DeleteVolumes(groupName, resp)
	err := <-resp
	if err != nil {
		log.Printf("[%08X] warning: release volumes fail: %s", id, err.Error())
	}
}

func (executor *CreateInstanceExecutor) ReleaseNetworkResource(id framework.SessionID, instance string) {
	resp := make(chan error)
	executor.NetworkModule.DeallocateAllResource(instance, resp)
	err := <-resp
	if err != nil {
		log.Printf("[%08X] warning: release network fail: %s", id, err.Error())
	}
}

func (executor *CreateInstanceExecutor) startAutoStartInstance(id framework.SessionID, instanceID, instanceName, receiver string) {
	var respChan = make(chan error, 1)
	executor.InstanceModule.StartInstance(instanceID, respChan)
	var err = <-respChan
	if err != nil {
		log.Printf("[%08X] warning: start autostart instance '%s' fail: %s", id, instanceName, err.Error())
		return
	}
	log.Printf("[%08X] autostart instance '%s' started", id, instanceName)

	event, _ := framework.CreateJsonMessage(framework.GuestStartedEvent)
	event.SetFromSession(id)
	event.SetString(framework.ParamKeyInstance, instanceID)
	if err = executor.Sender.SendMessage(event, receiver); err != nil {
		log.Printf("[%08X] notify guest started to '%s' fail: %s", id, receiver, err.Error())
	}
}

func (executor *CreateInstanceExecutor) generateMacAddress() (string, error) {
	const (
		BufferSize = 3
		MacPrefix  = "00:16:3e"
	)
	buf := make([]byte, BufferSize)
	_, err := executor.RandomGenerator.Read(buf)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%02x:%02x:%02x", MacPrefix, buf[0], buf[1], buf[2]), nil
}

func (executor *CreateInstanceExecutor) generatePassword(length int) string {
	const (
		Letters = "~!@#$%^&*()_[]-=+0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	)
	var result = make([]byte, length)
	var n = len(Letters)
	for i := 0; i < length; i++ {
		result[i] = Letters[executor.RandomGenerator.Intn(n)]
	}
	return string(result)
}

func (executor *CreateInstanceExecutor) ResponseFail(resp framework.Message, err, target string) error {
	resp.SetSuccess(false)
	resp.SetError(err)
	return executor.Sender.SendMessage(resp, target)
}
