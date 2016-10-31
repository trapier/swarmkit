package networkallocator

import (
	"testing"

	"github.com/docker/swarmkit/api"
	"github.com/stretchr/testify/assert"
)

func TestReconcilePortConfigs(t *testing.T) {
	type portConfigsBind struct {
		input  *api.Service
		expect []*api.PortConfig
	}

	portConfigsBinds := []portConfigsBind{
		{
			input: &api.Service{
				Spec: api.ServiceSpec{
					Endpoint: &api.EndpointSpec{
						Ports: []*api.PortConfig{
							{
								Name:          "test1",
								Protocol:      api.ProtocolTCP,
								TargetPort:    10000,
								PublishedPort: 10000,
							},
						},
					},
				},
				Endpoint: nil,
			},
			expect: []*api.PortConfig{
				{
					Name:          "test1",
					Protocol:      api.ProtocolTCP,
					TargetPort:    10000,
					PublishedPort: 10000,
				},
			},
		},
		{
			input: &api.Service{
				Spec: api.ServiceSpec{
					Endpoint: &api.EndpointSpec{
						Ports: []*api.PortConfig{
							{
								Name:          "test1",
								Protocol:      api.ProtocolTCP,
								TargetPort:    10000,
								PublishedPort: 10000,
							},
							{
								Name:          "test2",
								Protocol:      api.ProtocolTCP,
								TargetPort:    10001,
								PublishedPort: 10001,
							},
						},
					},
				},
				Endpoint: &api.Endpoint{
					Ports: []*api.PortConfig{
						{
							Name:          "test1",
							Protocol:      api.ProtocolTCP,
							TargetPort:    10000,
							PublishedPort: 10000,
						},
					},
				},
			},
			expect: []*api.PortConfig{
				{
					Name:          "test1",
					Protocol:      api.ProtocolTCP,
					TargetPort:    10000,
					PublishedPort: 10000,
				},
				{
					Name:          "test2",
					Protocol:      api.ProtocolTCP,
					TargetPort:    10001,
					PublishedPort: 10001,
				},
			},
		},
		{
			input: &api.Service{
				Spec: api.ServiceSpec{
					Endpoint: &api.EndpointSpec{
						Ports: []*api.PortConfig{
							{
								Name:          "test1",
								Protocol:      api.ProtocolTCP,
								TargetPort:    10000,
								PublishedPort: 10000,
							},
						},
					},
				},
				Endpoint: &api.Endpoint{
					Ports: []*api.PortConfig{
						{
							Name:          "test2",
							Protocol:      api.ProtocolTCP,
							TargetPort:    10001,
							PublishedPort: 10000,
						},
					},
				},
			},
			expect: []*api.PortConfig{
				{
					Name:          "test1",
					Protocol:      api.ProtocolTCP,
					TargetPort:    10000,
					PublishedPort: 10000,
				},
			},
		},
		{
			input: &api.Service{
				Spec: api.ServiceSpec{
					Endpoint: &api.EndpointSpec{
						Ports: []*api.PortConfig{
							{
								Name:          "test1",
								Protocol:      api.ProtocolTCP,
								TargetPort:    10000,
								PublishedPort: 0,
							},
							{
								Name:          "test2",
								Protocol:      api.ProtocolTCP,
								TargetPort:    10001,
								PublishedPort: 0,
							},
						},
					},
				},
				Endpoint: &api.Endpoint{
					Ports: []*api.PortConfig{
						{
							Name:          "test2",
							Protocol:      api.ProtocolTCP,
							TargetPort:    10001,
							PublishedPort: 10001,
						},
						{
							Name:          "test1",
							Protocol:      api.ProtocolTCP,
							TargetPort:    10000,
							PublishedPort: 10000,
						},
					},
				},
			},
			expect: []*api.PortConfig{
				{
					Name:          "test1",
					Protocol:      api.ProtocolTCP,
					TargetPort:    10000,
					PublishedPort: 10000,
				},
				{
					Name:          "test2",
					Protocol:      api.ProtocolTCP,
					TargetPort:    10001,
					PublishedPort: 10001,
				},
			},
		},
	}

	for _, singleTest := range portConfigsBinds {
		expect := reconcilePortConfigs(singleTest.input)
		assert.Equal(t, singleTest.expect, expect)
	}
}

func TestServiceAllocatePorts(t *testing.T) {
	pa, err := newPortAllocator()
	assert.NoError(t, err)

	// Service has no endpoint in ServiceSpec
	s := &api.Service{
		Spec: api.ServiceSpec{
			Endpoint: nil,
		},
		Endpoint: &api.Endpoint{
			Ports: []*api.PortConfig{
				{
					Name:          "test1",
					Protocol:      api.ProtocolTCP,
					TargetPort:    10000,
					PublishedPort: 10000,
				},
			},
		},
	}

	err = pa.serviceAllocatePorts(s)
	assert.NoError(t, err)

	// Service has a published port 10001 in ServiceSpec
	s = &api.Service{
		Spec: api.ServiceSpec{
			Endpoint: &api.EndpointSpec{
				Ports: []*api.PortConfig{
					{
						Name:          "test1",
						Protocol:      api.ProtocolTCP,
						TargetPort:    10000,
						PublishedPort: 10001,
					},
				},
			},
		},
		Endpoint: &api.Endpoint{
			Ports: []*api.PortConfig{
				{
					Name:          "test1",
					Protocol:      api.ProtocolTCP,
					TargetPort:    10000,
					PublishedPort: 10000,
				},
			},
		},
	}

	err = pa.serviceAllocatePorts(s)
	assert.NoError(t, err)

	// Service has a published port 10001 in ServiceSpec
	// which is already allocated on host
	s = &api.Service{
		Spec: api.ServiceSpec{
			Endpoint: &api.EndpointSpec{
				Ports: []*api.PortConfig{
					{
						Name:          "test1",
						Protocol:      api.ProtocolTCP,
						TargetPort:    10000,
						PublishedPort: 10001,
					},
				},
			},
		},
		Endpoint: &api.Endpoint{
			Ports: []*api.PortConfig{
				{
					Name:          "test1",
					Protocol:      api.ProtocolTCP,
					TargetPort:    10000,
					PublishedPort: 10000,
				},
			},
		},
	}

	// port allocated already, got an error
	err = pa.serviceAllocatePorts(s)
	assert.Error(t, err)
}

func TestIsPortsAllocated(t *testing.T) {
	pa, err := newPortAllocator()
	assert.NoError(t, err)

	type Data struct {
		input  *api.Service
		expect bool
	}

	testCases := []Data{
		{
			// both Endpoint and Spec.Endpoint are nil
			input: &api.Service{
				Spec: api.ServiceSpec{
					Endpoint: nil,
				},
				Endpoint: nil,
			},
			expect: true,
		},
		{
			// Endpoint is non-nil and Spec.Endpoint is nil
			input: &api.Service{
				Spec: api.ServiceSpec{
					Endpoint: &api.EndpointSpec{
						Ports: []*api.PortConfig{
							{
								Name:          "test1",
								Protocol:      api.ProtocolTCP,
								TargetPort:    10000,
								PublishedPort: 10000,
							},
						},
					},
				},
				Endpoint: nil,
			},
			expect: false,
		},
		{
			// Endpoint is nil and Spec.Endpoint is non-nil
			input: &api.Service{
				Spec: api.ServiceSpec{
					Endpoint: nil,
				},
				Endpoint: &api.Endpoint{
					Ports: []*api.PortConfig{
						{
							Name:          "test2",
							Protocol:      api.ProtocolTCP,
							TargetPort:    10001,
							PublishedPort: 10000,
						},
					},
				},
			},
			expect: false,
		},
		{
			// Endpoint and Spec.Endpoint have different length
			input: &api.Service{
				Spec: api.ServiceSpec{
					Endpoint: &api.EndpointSpec{
						Ports: []*api.PortConfig{
							{
								Name:          "test1",
								Protocol:      api.ProtocolTCP,
								TargetPort:    10000,
								PublishedPort: 10000,
							},
							{
								Name:          "test2",
								Protocol:      api.ProtocolTCP,
								TargetPort:    10001,
								PublishedPort: 10001,
							},
						},
					},
				},
				Endpoint: &api.Endpoint{
					Ports: []*api.PortConfig{
						{
							Name:          "test2",
							Protocol:      api.ProtocolTCP,
							TargetPort:    10001,
							PublishedPort: 10000,
						},
					},
				},
			},
			expect: false,
		},
		{
			// Endpoint and Spec.Endpoint have different TargetPort
			input: &api.Service{
				Spec: api.ServiceSpec{
					Endpoint: &api.EndpointSpec{
						Ports: []*api.PortConfig{
							{
								Name:          "test1",
								Protocol:      api.ProtocolTCP,
								TargetPort:    10001,
								PublishedPort: 10000,
							},
						},
					},
				},
				Endpoint: &api.Endpoint{
					Ports: []*api.PortConfig{
						{
							Name:          "test1",
							Protocol:      api.ProtocolTCP,
							TargetPort:    10000,
							PublishedPort: 10000,
						},
					},
				},
			},
			expect: false,
		},
		{
			// Endpoint and Spec.Endpoint have different PublishedPort
			input: &api.Service{
				Spec: api.ServiceSpec{
					Endpoint: &api.EndpointSpec{
						Ports: []*api.PortConfig{
							{
								Name:          "test1",
								Protocol:      api.ProtocolTCP,
								TargetPort:    10000,
								PublishedPort: 10001,
							},
						},
					},
				},
				Endpoint: &api.Endpoint{
					Ports: []*api.PortConfig{
						{
							Name:          "test1",
							Protocol:      api.ProtocolTCP,
							TargetPort:    10000,
							PublishedPort: 10000,
						},
					},
				},
			},
			expect: false,
		},
		{
			// Endpoint and Spec.Endpoint are the same and PublishedPort is 0
			input: &api.Service{
				Spec: api.ServiceSpec{
					Endpoint: &api.EndpointSpec{
						Ports: []*api.PortConfig{
							{
								Name:          "test1",
								Protocol:      api.ProtocolTCP,
								TargetPort:    10000,
								PublishedPort: 0,
							},
						},
					},
				},
				Endpoint: &api.Endpoint{
					Ports: []*api.PortConfig{
						{
							Name:          "test1",
							Protocol:      api.ProtocolTCP,
							TargetPort:    10000,
							PublishedPort: 0,
						},
					},
				},
			},
			expect: false,
		},
		{
			// Endpoint and Spec.Endpoint are the same and PublishedPort is non-0
			input: &api.Service{
				Spec: api.ServiceSpec{
					Endpoint: &api.EndpointSpec{
						Ports: []*api.PortConfig{
							{
								Name:          "test1",
								Protocol:      api.ProtocolTCP,
								TargetPort:    10000,
								PublishedPort: 10000,
							},
						},
					},
				},
				Endpoint: &api.Endpoint{
					Ports: []*api.PortConfig{
						{
							Name:          "test1",
							Protocol:      api.ProtocolTCP,
							TargetPort:    10000,
							PublishedPort: 10000,
						},
					},
				},
			},
			expect: true,
		},
		{
			// Endpoint and Spec.Endpoint are the same except PublishedPort, and PublishedPort in Endpoint is non-0
			input: &api.Service{
				Spec: api.ServiceSpec{
					Endpoint: &api.EndpointSpec{
						Ports: []*api.PortConfig{
							{
								Name:          "test1",
								Protocol:      api.ProtocolTCP,
								TargetPort:    10000,
								PublishedPort: 0,
							},
						},
					},
				},
				Endpoint: &api.Endpoint{
					Ports: []*api.PortConfig{
						{
							Name:          "test1",
							Protocol:      api.ProtocolTCP,
							TargetPort:    10000,
							PublishedPort: 10000,
						},
					},
				},
			},
			expect: true,
		},
		{
			// Endpoint and Spec.Endpoint are the same except the ports are in different order
			input: &api.Service{
				Spec: api.ServiceSpec{
					Endpoint: &api.EndpointSpec{
						Ports: []*api.PortConfig{
							{
								Name:          "test1",
								Protocol:      api.ProtocolTCP,
								TargetPort:    10000,
								PublishedPort: 0,
							},
							{
								Name:          "test2",
								Protocol:      api.ProtocolTCP,
								TargetPort:    10001,
								PublishedPort: 0,
							},
							{
								Name:          "test3",
								Protocol:      api.ProtocolTCP,
								TargetPort:    10002,
								PublishedPort: 0,
								PublishMode:   api.PublishModeHost,
							},
						},
					},
				},
				Endpoint: &api.Endpoint{
					Ports: []*api.PortConfig{
						{
							Name:          "test2",
							Protocol:      api.ProtocolTCP,
							TargetPort:    10001,
							PublishedPort: 10001,
						},
						{
							Name:          "test3",
							Protocol:      api.ProtocolTCP,
							TargetPort:    10002,
							PublishedPort: 0,
							PublishMode:   api.PublishModeHost,
						},
						{
							Name:          "test1",
							Protocol:      api.ProtocolTCP,
							TargetPort:    10000,
							PublishedPort: 10000,
						},
					},
				},
			},
			expect: true,
		},
	}

	for _, singleTest := range testCases {
		expect := pa.isPortsAllocated(singleTest.input)
		assert.Equal(t, expect, singleTest.expect)
	}
}

func TestAllocate(t *testing.T) {
	pSpace, err := newPortSpace(api.ProtocolTCP)
	assert.NoError(t, err)

	pConfig := &api.PortConfig{
		Name:          "test1",
		Protocol:      api.ProtocolTCP,
		TargetPort:    30000,
		PublishedPort: 30000,
	}

	// first consume 30000 in dynamicPortSpace
	err = pSpace.allocate(pConfig)
	assert.NoError(t, err)

	pConfig = &api.PortConfig{
		Name:          "test1",
		Protocol:      api.ProtocolTCP,
		TargetPort:    30000,
		PublishedPort: 30000,
	}

	// consume 30000 again in dynamicPortSpace, got an error
	err = pSpace.allocate(pConfig)
	assert.Error(t, err)

	pConfig = &api.PortConfig{
		Name:          "test2",
		Protocol:      api.ProtocolTCP,
		TargetPort:    30000,
		PublishedPort: 10000,
	}

	// consume 10000 in masterPortSpace, got no error
	err = pSpace.allocate(pConfig)
	assert.NoError(t, err)

	pConfig = &api.PortConfig{
		Name:          "test3",
		Protocol:      api.ProtocolTCP,
		TargetPort:    30000,
		PublishedPort: 10000,
	}

	// consume 10000 again in masterPortSpace, got an error
	err = pSpace.allocate(pConfig)
	assert.Error(t, err)
}
