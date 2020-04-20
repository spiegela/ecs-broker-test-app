package main

type bindingLookup map[string][]serviceBinding

type serviceBinding struct {
	BindingName    string             `json:"binding_name"`
	InstanceName   string             `json:"instance_name"`
	Label          string             `json:"label"`
	Name           string             `json:"name"`
	Plan           string             `json:"plan"`
	Provider       string             `json:"provider"`
	SyslogDrainURL string             `json:"syslog_drain_url"`
	Tags           []string           `json:"tags"`
	VolumeMounts   []volumeMount      `json:"volume_mounts"`
	Credentials    serviceCredentials `json:"credentials"`
}

type serviceCredentials struct {
	AccessKey       string `json:"accessKey"`
	Bucket          string `json:"bucket"`
	Endpoint        string `json:"endpoint"`
	PathStyleAccess bool   `json:"path-style-access"`
	S3URL           string `json:"s3Url"`
	SecretKey       string `json:"secretKey"`
}

type volumeMount struct {
	Driver       string `json:"driver"`
	ContainerDir string `json:"container_dir"`
	Mode         string `json:"mode"`
	DeviceType   string `json:"device_type"`
	Device       device `json:"device"`
}

type device struct {
	VolumeID    string                 `json:"volume_id"`
	MountConfig map[string]interface{} `json"mount_config"`
}
