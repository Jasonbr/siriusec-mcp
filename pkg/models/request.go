/*
请求模型定义

定义所有MCP工具使用的请求结构体
*/
package models

// MCPRequest MCP请求基类
type MCPRequest struct {
	// 基础字段，可根据需要扩展
}

// ListAllInstancesRequest 列出所有实例请求
type ListAllInstancesRequest struct {
	MCPRequest
	Region       string `json:"region,omitempty"`
	ManagedType  string `json:"managedType,omitempty"`
	InstanceType string `json:"instanceType,omitempty"`
	PluginID     string `json:"pluginId,omitempty"`
	Filters      string `json:"filters,omitempty"`
	Current      string `json:"current,omitempty"`
	PageSize     string `json:"pageSize,omitempty"`
	MaxResults   int    `json:"maxResults,omitempty"`
	NextToken    string `json:"nextToken,omitempty"`
}

// ListInstancesRequest 列出实例请求
type ListInstancesRequest struct {
	MCPRequest
	Instance  string `json:"instance,omitempty"`
	Status    string `json:"status,omitempty"`
	Region    string `json:"region,omitempty"`
	ClusterID string `json:"clusterId,omitempty"`
	Current   int    `json:"current,omitempty"`
	PageSize  int    `json:"pageSize,omitempty"`
}

// ListClustersRequest 列出集群请求
type ListClustersRequest struct {
	MCPRequest
	Name          string `json:"name,omitempty"`
	ClusterID     string `json:"clusterId,omitempty"`
	ClusterType   string `json:"clusterType,omitempty"`
	ClusterStatus string `json:"clusterStatus,omitempty"`
	Current       int    `json:"current,omitempty"`
	PageSize      int    `json:"pageSize,omitempty"`
}

// ListPodsOfInstanceRequest 列出实例下Pod请求
type ListPodsOfInstanceRequest struct {
	MCPRequest
	Instance  string `json:"instance"`
	ClusterID string `json:"clusterId,omitempty"`
	Current   int    `json:"current,omitempty"`
	PageSize  int    `json:"pageSize,omitempty"`
}

// DiagnosisRequestParams 诊断请求参数基类
type DiagnosisRequestParams struct {
	Region string `json:"region"`
	Hide   string `json:"_hide,omitempty"`
}

// DiagnosisRequest 诊断请求
type DiagnosisRequest struct {
	MCPRequest
	ServiceName string                 `json:"service_name"`
	Channel     string                 `json:"channel"`
	Region      string                 `json:"region"`
	Params      map[string]interface{} `json:"params"`
}

// MemGraphDiagnosisParams memgraph诊断参数
type MemGraphDiagnosisParams struct {
	DiagnosisRequestParams
	Instance    string `json:"instance,omitempty"`
	Pod         string `json:"pod,omitempty"`
	ClusterType string `json:"clusterType,omitempty"`
	ClusterID   string `json:"clusterId,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
}

// JavaMemDiagnosisParams javamem诊断参数
type JavaMemDiagnosisParams struct {
	DiagnosisRequestParams
	Instance    string `json:"instance"`
	Pid         string `json:"Pid,omitempty"`
	Pod         string `json:"pod,omitempty"`
	Duration    string `json:"duration,omitempty"`
	ClusterType string `json:"clusterType,omitempty"`
	ClusterID   string `json:"clusterId,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
}

// OOMCheckDiagnosisParams oomcheck诊断参数
type OOMCheckDiagnosisParams struct {
	DiagnosisRequestParams
	Instance    string `json:"instance,omitempty"`
	Pod         string `json:"pod,omitempty"`
	Time        string `json:"time,omitempty"`
	ClusterType string `json:"clusterType,omitempty"`
	ClusterID   string `json:"clusterId,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
}

// IODiagnosisParams IO诊断参数
type IODiagnosisParams struct {
	DiagnosisRequestParams
	Instance  string `json:"instance,omitempty"`
	Pod       string `json:"pod,omitempty"`
	Pid       string `json:"pid,omitempty"`
	Duration  string `json:"duration,omitempty"`
	ClusterID string `json:"clusterId,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// NetDiagnosisParams 网络诊断参数
type NetDiagnosisParams struct {
	DiagnosisRequestParams
	Instance    string `json:"instance,omitempty"`
	Pod         string `json:"pod,omitempty"`
	ClusterType string `json:"clusterType,omitempty"`
	ClusterID   string `json:"clusterId,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	Protocol    string `json:"protocol,omitempty"`
	SrcPort     string `json:"srcPort,omitempty"`
	DstPort     string `json:"dstPort,omitempty"`
	SrcIP       string `json:"srcIP,omitempty"`
	DstIP       string `json:"dstIP,omitempty"`
}

// SchedDiagnosisParams 调度诊断参数
type SchedDiagnosisParams struct {
	DiagnosisRequestParams
	Instance  string `json:"instance,omitempty"`
	Pod       string `json:"pod,omitempty"`
	ClusterID string `json:"clusterId,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Pid       string `json:"pid,omitempty"`
}

// CrashAgentRequest 崩溃诊断代理请求
type CrashAgentRequest struct {
	MCPRequest
	Action    string `json:"action"`
	Instance  string `json:"instance,omitempty"`
	ClusterID string `json:"clusterId,omitempty"`
	Region    string `json:"region,omitempty"`
}

// InitialSysomRequest 初始化Sysom请求
type InitialSysomRequest struct {
	MCPRequest
	UID    string `json:"uid"`
	Region string `json:"region,omitempty"`
}

// CheckSysomInitialedRequest 检查Sysom初始化状态请求
type CheckSysomInitialedRequest struct {
	MCPRequest
	UID    string `json:"uid"`
	Region string `json:"region,omitempty"`
}
