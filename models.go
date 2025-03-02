package main

// AuthResponse представляет ответ аутентификации.
type AuthResponse struct {
    RefreshToken string `json:"refresh_token"`
}

// Options содержит параметры конфигурации кластера.
type Options struct {
    MaximumLagOnFailover  int  `json:"maximum_lag_on_failover"`
    WalArchiveMode        bool `json:"wal_archive_mode"`
    AutoRestart           bool `json:"auto_restart"`
    Production            bool `json:"production"`
    EnableSynchronousMode bool `json:"enable_synchronous_mode"`
    DisableAutofailover   bool `json:"disable_autofailover"`
}

// CreateClusterRequest представляет запрос на создание кластера.
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

// Instance представляет экземпляр кластера.
type Instance struct {
    ClusterID string `json:"cluster_id"`
}

// CreateClusterResponse представляет ответ на запрос создания кластера.
type CreateClusterResponse struct {
    Instances []Instance `json:"instances"`
}

// TableSpaceResponse представляет ответ с информацией о таблице.
type TableSpaceResponse struct {
    Id string `json:"id"`
}

// CreateDBRequest представляет запрос на создание базы данных.
type CreateDBRequest struct {
    Name         string `json:"name"`
    TableSpaceID string `json:"tablespace_id"`
}

// CreateDBResponse представляет ответ на запрос создания базы данных.
type CreateDBResponse struct {
    Id string `json:"id"`
}

// CreateClusterUserRequest представляет запрос на создание пользователя кластера.
type CreateClusterUserRequest struct {
    Databases []string `json:"databases"`
    Roles     []string `json:"roles"`
    Name      string   `json:"name"`
    Password  string   `json:"password"`
}

// ResponseDBUsers представляет ответ с информацией о пользователях базы данных.
type ResponseDBUsers struct {
    MasterConnectionString string `json:"master_connection_string"`
}

// ClusterStatusResponse представляет ответ с информацией о статусе кластера.
type ClusterStatusResponse struct {
    Status string `json:"status"`
}

// dbStatusResponse представляет глобальную переменную с информацией о статусе базы данных.
var dbStatusResponse struct {
    Status string `json:"status"`
}