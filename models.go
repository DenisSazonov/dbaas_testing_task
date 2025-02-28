package main

type AuthResponse struct {
    RefreshToken string `json:"refresh_token"`
}

type Options struct {
    MaximumLagOnFailover  int  `json:"maximum_lag_on_failover"`
    WalArchiveMode        bool `json:"wal_archive_mode"`
    AutoRestart           bool `json:"auto_restart"`
    Production            bool `json:"production"`
    EnableSynchronousMode bool `json:"enable_synchronous_mode"`
    DisableAutofailover   bool `json:"disable_autofailover"`
}

type CreateClusterRequest struct {
    TypeID        string  `json:"type_id"`
    Options       Options `json:"options"`
    DiskSize      int64   `json:"disk_size"`
    Mode          string  `json:"mode"`
    ReplicasCount int     `json:"replicas_count"`
    CreationMode  string  `json:"creation_mode"`
    Name          string  `json:"name"`
    FlavorID      string  `json:"flavor_id"`
    TypeName      string  `json:"type_name"`
    Az            string  `json:"az"`
    HAManager     string  `json:"ha_manager"`
    HA            bool    `json:"ha"`
}

type Instance struct {
    ClusterID string `json:"cluster_id"`
}

type CreateClusterResponse struct {
    Instances []Instance `json:"instances"`
}

type TableSpaceResponse struct {
    Id string `json:"id"`
}

type CreateDBRequest struct {
    Name         string `json:"name"`
    TableSpaceID string `json:"tablespace_id"`
}

type CreateDBResponse struct {
    Id string `json:"id"`
}

type CreateClusterUserRequest struct {
    Databases []string `json:"databases"`
    Roles     []string `json:"roles"`
    Name      string   `json:"name"`
    Password  string   `json:"password"`
}

type ResponseDBUsers struct {
    MasterConnectionString string `json:"master_connection_string"`
}

type ClusterStatusResponse struct {
	Status string `json:"status"`
}

var dbStatusResponse struct {
	Status string `json:"status"`
}