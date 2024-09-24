/*
Copyright 2021 Alibaba Group Holding Limited.

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

package factory

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	polardbxv1reconcile "github.com/alibaba/polardbx-operator/pkg/operator/v1/polardbx/reconcile"
	"github.com/alibaba/polardbx-operator/pkg/probe"
)

type ProbeConfigure interface {
	ConfigureForCNEngine(container *corev1.Container, ports CNPorts)
	ConfigureForCNExporter(container *corev1.Container, ports CNPorts)
	ConfigureForCDCEngine(container *corev1.Container, ports CDCPorts)
	ConfigureForCDCExporter(container *corev1.Container, ports CDCPorts)
}

type probeConfigure struct {
	rc       *polardbxv1reconcile.Context
	polardbx *polardbxv1.PolarDBXCluster
}

func (p *probeConfigure) newProbeWithProber(endpoint string, probeTarget string, ports ProberPort) corev1.ProbeHandler {
	return corev1.ProbeHandler{
		HTTPGet: &corev1.HTTPGetAction{
			Path: endpoint,
			Port: intstr.FromInt(ports.GetProbePort()),
			HTTPHeaders: []corev1.HTTPHeader{
				{Name: "Probe-Target", Value: probeTarget},
				{Name: "Probe-Port", Value: strconv.Itoa(ports.GetAccessPort())},
				{Name: "Probe-Timeout", Value: "10s"},
			},
		},
	}
}

func (p *probeConfigure) ConfigureForCNEngine(container *corev1.Container, ports CNPorts) {
	container.StartupProbe = &corev1.Probe{
		InitialDelaySeconds: 10,
		TimeoutSeconds:      10,
		PeriodSeconds:       10,
		FailureThreshold:    300,
		ProbeHandler:        p.newProbeWithProber("/liveness", probe.TypePolarDBX, &ports),
	}
	container.LivenessProbe = &corev1.Probe{
		TimeoutSeconds: 10,
		PeriodSeconds:  10,
		ProbeHandler:   p.newProbeWithProber("/liveness", probe.TypePolarDBX, &ports),
	}
	container.ReadinessProbe = &corev1.Probe{
		TimeoutSeconds: 10,
		PeriodSeconds:  10,
		ProbeHandler:   p.newProbeWithProber("/readiness", probe.TypePolarDBX, &ports),
	}
}

func (p *probeConfigure) ConfigureForCNExporter(container *corev1.Container, ports CNPorts) {
	container.LivenessProbe = &corev1.Probe{
		TimeoutSeconds: 5,
		PeriodSeconds:  20,
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt(ports.MetricsPort),
			},
		},
	}
	container.ReadinessProbe = &corev1.Probe{
		TimeoutSeconds: 5,
		PeriodSeconds:  20,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/metrics",
				Port: intstr.FromInt(ports.MetricsPort),
			},
		},
	}
}

func (p *probeConfigure) ConfigureForCDCEngine(container *corev1.Container, ports CDCPorts) {
	hanlder := corev1.ProbeHandler{
		TCPSocket: &corev1.TCPSocketAction{
			Port: intstr.FromInt(ports.GetAccessPort()),
		},
	}
	container.StartupProbe = &corev1.Probe{
		InitialDelaySeconds: 10,
		TimeoutSeconds:      10,
		PeriodSeconds:       10,
		FailureThreshold:    18,
		ProbeHandler:        hanlder,
	}
	container.LivenessProbe = &corev1.Probe{
		TimeoutSeconds:   10,
		PeriodSeconds:    10,
		FailureThreshold: 5,
		ProbeHandler:     hanlder,
	}
}

func (p *probeConfigure) ConfigureForCDCExporter(container *corev1.Container, ports CDCPorts) {
	container.LivenessProbe = &corev1.Probe{
		TimeoutSeconds: 5,
		PeriodSeconds:  20,
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt(ports.MetricsPort),
			},
		},
	}
	container.ReadinessProbe = &corev1.Probe{
		TimeoutSeconds: 5,
		PeriodSeconds:  20,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/metrics",
				Port: intstr.FromInt(ports.MetricsPort),
			},
		},
	}
}

func NewProbeConfigure(rc *polardbxv1reconcile.Context, pxc *polardbxv1.PolarDBXCluster) ProbeConfigure {
	return &probeConfigure{
		rc:       rc,
		polardbx: pxc,
	}
}
