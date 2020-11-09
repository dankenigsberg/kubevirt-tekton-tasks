package vmcreator

import (
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/create-vm/pkg/datavolume"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/create-vm/pkg/pvc"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/create-vm/pkg/templates"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/create-vm/pkg/templates/validations"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/create-vm/pkg/utils/log"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/create-vm/pkg/utils/parse"
	virtualMachine "github.com/kubevirt/kubevirt-tekton-tasks/modules/create-vm/pkg/vm"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/shared/pkg/env"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/shared/pkg/zconstants"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/shared/pkg/zerrors"
	templatev1 "github.com/openshift/client-go/template/clientset/versioned/typed/template/v1"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
	kubevirtcliv1 "kubevirt.io/client-go/kubecli"
	datavolumeclientv1alpha1 "kubevirt.io/containerized-data-importer/pkg/client/clientset/versioned/typed/core/v1alpha1"
	"path/filepath"
)

type VMCreator struct {
	targetNamespace        string
	cliOptions             *parse.CLIOptions
	config                 *rest.Config
	templateProvider       templates.TemplateProvider
	virtualMachineProvider virtualMachine.VirtualMachineProvider
	dataVolumeProvider     datavolume.DataVolumeProvider
	pvcProvider            pvc.PersistentVolumeClaimProvider
}

func getConfig() (*rest.Config, error) {
	if env.IsEnvVarTrue(zconstants.OutOfClusterENV) {
		return clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	}
	return rest.InClusterConfig()
}

func NewVMCreator(cliOptions *parse.CLIOptions) (*VMCreator, error) {
	log.GetLogger().Debug("initialized clients and providers")
	targetNS := cliOptions.GetVirtualMachineNamespace()

	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	// clients
	kubeClient := kubernetes.NewForConfigOrDie(config)
	templateClient := templatev1.NewForConfigOrDie(config)
	cdiClient := datavolumeclientv1alpha1.NewForConfigOrDie(config)
	kubevirtClient, err := kubevirtcliv1.GetKubevirtClientFromRESTConfig(config)
	if err != nil {
		return nil, errors.WithMessage(err, "Cannot create kubevirt client")
	}

	templateProvider := templates.NewTemplateProvider(templateClient)
	virtualMachineProvider := virtualMachine.NewVirtualMachineProvider(kubevirtClient)
	dataVolumeProvider := datavolume.NewDataVolumeProvider(cdiClient)
	pvcProvider := pvc.NewPersistentVolumeClaimProvider(kubeClient.CoreV1())

	return &VMCreator{
		targetNamespace:        targetNS,
		cliOptions:             cliOptions,
		config:                 config,
		templateProvider:       templateProvider,
		virtualMachineProvider: virtualMachineProvider,
		dataVolumeProvider:     dataVolumeProvider,
		pvcProvider:            pvcProvider,
	}, nil
}

func (v *VMCreator) CreateVM() (*kubevirtv1.VirtualMachine, error) {
	log.GetLogger().Debug("retrieving template", zap.String("name", v.cliOptions.TemplateName), zap.String("namespace", v.cliOptions.GetTemplateNamespace()))
	template, err := v.templateProvider.Get(v.cliOptions.GetTemplateNamespace(), v.cliOptions.TemplateName)
	if err != nil {
		return nil, err
	}

	log.GetLogger().Debug("processing template", zap.String("name", v.cliOptions.TemplateName), zap.String("namespace", v.cliOptions.GetTemplateNamespace()))
	processedTemplate, err := v.templateProvider.Process(v.targetNamespace, template, v.cliOptions.GetTemplateParams())
	if err != nil {
		return nil, err
	}
	vm, err := templates.DecodeVM(processedTemplate)
	if err != nil {
		return nil, err
	}

	templateValidations, err := templates.GetTemplateValidations(processedTemplate)
	if err != nil {
		log.GetLogger().Warn("could not parse template validations", zap.Error(err))
		templateValidations = validations.NewTemplateValidations(nil) // fallback to defaults
	}
	if templateValidations.IsEmpty() {
		log.GetLogger().Debug("template validations are empty: falling back to defaults")
	}

	vm.Namespace = v.targetNamespace
	virtualMachine.AddMetadata(vm, processedTemplate)
	virtualMachine.AddVolumes(vm, templateValidations, v.cliOptions)

	log.GetLogger().Debug("creating VM", zap.Reflect("vm", vm))
	return v.virtualMachineProvider.Create(v.targetNamespace, vm)
}

func (v *VMCreator) CheckVolumesExist() error {
	log.GetLogger().Debug("asserting additional volumes exist", zap.Strings("additional-volumes", v.cliOptions.GetAllDiskNames()))
	_, dvsErr := v.dataVolumeProvider.GetByName(v.targetNamespace, v.cliOptions.GetAllDVNames()...)
	_, pvcsErr := v.pvcProvider.GetByName(v.targetNamespace, v.cliOptions.GetAllPVCNames()...)

	return zerrors.NewMultiError().
		AddC("dvsErr", dvsErr).
		AddC("pvcsErr", pvcsErr).
		AsOptional()
}

func (v *VMCreator) OwnVolumes(vm *kubevirtv1.VirtualMachine) error {
	dvsErr := v.ownDataVolumes(vm)
	pvcsErr := v.ownPersistentVolumeClaims(vm)

	return zerrors.NewMultiError().
		AddC("dvsErr", dvsErr).
		AddC("pvcsErr", pvcsErr).
		AsOptional()
}

func (v *VMCreator) ownDataVolumes(vm *kubevirtv1.VirtualMachine) error {
	log.GetLogger().Debug("taking ownership of DataVolumes", zap.Strings("own-dvs", v.cliOptions.OwnDataVolumes))
	var multiError zerrors.MultiError

	dvs, dvsErr := v.dataVolumeProvider.GetByName(v.targetNamespace, v.cliOptions.OwnDataVolumes...)

	for idx, dvName := range v.cliOptions.OwnDataVolumes {
		if err := zerrors.GetErrorFromMultiError(dvsErr, dvName); err != nil {
			multiError.Add(dvName, err)
			continue
		}

		if _, err := v.dataVolumeProvider.AddOwnerReferences(dvs[idx], virtualMachine.AsVMOwnerReference(vm)); err != nil {
			multiError.Add(dvName, errors.Wrapf(err, "could not add owner reference to %v DataVolume", dvName))
		}

	}

	return multiError.AsOptional()
}

func (v *VMCreator) ownPersistentVolumeClaims(vm *kubevirtv1.VirtualMachine) error {
	log.GetLogger().Debug("taking ownership of PersistentVolumeClaims", zap.Strings("own-pvcs", v.cliOptions.OwnPersistentVolumeClaims))
	var multiError zerrors.MultiError

	pvcs, pvcsErr := v.pvcProvider.GetByName(v.targetNamespace, v.cliOptions.OwnPersistentVolumeClaims...)

	for idx, pvcName := range v.cliOptions.OwnPersistentVolumeClaims {
		if err := zerrors.GetErrorFromMultiError(pvcsErr, pvcName); err != nil {
			multiError.Add(pvcName, err)
			continue
		}

		if _, err := v.pvcProvider.AddOwnerReferences(pvcs[idx], virtualMachine.AsVMOwnerReference(vm)); err != nil {
			multiError.Add(pvcName, errors.Wrapf(err, "could not add owner reference to %v PersistentVolumeClaim", pvcName))
		}

	}

	return multiError.AsOptional()
}
