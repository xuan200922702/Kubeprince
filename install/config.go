package install

import (
	"bytes"
	"github.com/wonderivan/logger"
	"html/template"
	"strings"
)

const oneMBByte = 1024 * 1024

type PrinceInstaller struct {
	Hosts []string
}

var (
	Masters  []string
	Nodes    []string
	VIP            string
	PkgUrl         string
	User           string
	Password       string
	PrivateKeyFile string
	KubeadmFile    string
	LvsFile        string
	Version        string
	Kustomize      bool
	ApiServer      string
)



var (
	JoinToken       string
	TokenCaCertHash string
	CertificateKey  string
)

type CommandType string

const InitMaster CommandType = "initMaster"
const JoinMaster CommandType = "joinMaster"
const JoinNode CommandType = "joinNode"

const Templateyaml= string(`apiVersion: kubeadm.k8s.io/v1beta1
kind: ClusterConfiguration
kubernetesVersion: {{.Version}}
controlPlaneEndpoint: "{{.ApiServer}}:6443"
networking:
  podSubnet: 100.64.0.0/10
apiServer:
        certSANs:
        - 127.0.0.1
        - {{.ApiServer}}
        {{range .Masters -}}
        - {{.}}
        {{end -}}
        - {{.VIP}}
---
apiVersion: kubeproxy.config.k8s.io/v1alpha1
kind: KubeProxyConfiguration
mode: "ipvs"
ipvs:
        excludeCIDRs: 
        - "{{.VIP}}/32"`)

//var ConfigType string
//func Config() {
//	switch ConfigType {
//	case "kubeadm":
//		printlnKubeadmConfig()
//	default:
//		printlnKubeadmConfig()
//	}
//}
//func printlnKubeadmConfig() {
//	fmt.Println(kubeadmConfig())
//}


const  princelvsyaml CommandType =(`
apiVersion: v1
kind: Pod
metadata:
  labels:
    component: kubeprince-lvs
    tier: control-plane
  name: kubeprince-lvs
  namespace: kube-system
spec:
  containers:
  - command:
    - /usr/bin/lvscare
    - care
    - --vs
    - {{.VIP}}:6443
    - --health-path
    - /healthz
    - --health-schem
    - https
    - --rs
    {{range .Masters -}}
    - {{.}}:6443
    {{end -}}
    image: kubeprince-lvs:latest
    imagePullPolicy: IfNotPresent
    name: kubeprince-lvs
    securityContext:
      privileged: true
  hostNetwork: true
  priorityClassName: system-cluster-critical
status: {}`)

func kubeadmConfig() (string) {
	var sb strings.Builder
	sb.Write([]byte(Templateyaml))
	return sb.String()
}
func lvsConfig() (string) {
	var sb strings.Builder
	sb.Write([]byte(princelvsyaml))
	return sb.String()
}

func  Template()([]byte){
	return  TemplateFromTemplateContent(kubeadmConfig())
}

func  Template2()([]byte){
	return  TemplateFromTemplateContent(lvsConfig())
}

func TemplateFromTemplateContent(templateContent string) []byte {
	tmpl, err := template.New("text").Parse(templateContent)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("template parse failed:", err)
		}
	}()
	if err != nil {
		panic(1)
	}
	var masters []string
	for _, h := range Masters {
		masters = append(masters, IpFormat(h))
	}
	var envMap = make(map[string]interface{})
	envMap["VIP"] = VIP
	envMap["Masters"] = masters
	envMap["Version"] = Version
	envMap["ApiServer"] = ApiServer
	var buffer bytes.Buffer
	_ = tmpl.Execute(&buffer, envMap)
	return buffer.Bytes()
}
