package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	nacosLogger "github.com/nacos-group/nacos-sdk-go/common/logger"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

type ChangeListener func(namespace, group, dataId, data string)

type NacosProvider struct {
	ChangeListener ChangeListener
	NacosLogger    nacosLogger.Logger // custom logger for replacing nacos default logger
	LogLevel       string             // log level for nacos default logger
}

var _ Provider = &NacosProvider{}

func (p *NacosProvider) Name() string {
	return "nacos"
}

func (p *NacosProvider) Config(helper *providerHelper) ([]byte, error) {
	helper.log.Infow("begin create nacos client")
	if p.LogLevel == "" {
		p.LogLevel = "error"
	}
	client, err := p.newNacosClientFromEnv(helper.log)
	if err != nil {
		return nil, fmt.Errorf("newNacosClientFromEnv failed, err=%w", err)
	}

	helper.log.Infow("begin read config from nacos")
	nacosContent, err := client.readConfig()
	if err != nil {
		return nil, fmt.Errorf("read config from nacos failed, err=%w", err)
	}
	if nacosContent == "" {
		return nil, fmt.Errorf("read config from nacos failed, err=%w", ErrEmptyConfig)
	}
	helper.log.Infow("read config from nacos success")
	return []byte(nacosContent), nil
}

type NacosClient struct {
	client         config_client.IConfigClient
	servers        []string
	namespace      string
	group          string
	dataID         string
	log            Logger
	changeListener ChangeListener
	logLevel       string
	nacosLogger    nacosLogger.Logger
}

func (p *NacosProvider) newNacosClientFromEnv(log Logger) (*NacosClient, error) {
	// NACOS_HOST NACOS_PORT NACOS_NAMESPACE NACOS_GROUP NACOS_DATAID
	host := os.Getenv(EnvNacosHost)
	port := os.Getenv(EnvNacosPort)
	nacosServers := []string{fmt.Sprintf("%s:%s", host, port)}

	namespace := os.Getenv(EnvNacosNamespace)
	group := os.Getenv(EnvNacosGroup)
	dataID := os.Getenv(EnvNacosDataID)

	log.Debugw("reading nacos config from env", EnvNacosHost, host, EnvNacosPort, port, EnvNacosNamespace, namespace, EnvNacosGroup, group,
		EnvNacosDataID, dataID, "nacosServers", nacosServers)

	if namespace == "" || group == "" || dataID == "" || host == "" || port == "" {
		return nil, fmt.Errorf("no nacos config env vars found, abort loading. "+
			"%v=%v, %v=%v, %v=%v, %v=%v, %v=%v %v=%v",
			EnvNacosHost, host, EnvNacosPort, port, EnvNacosNamespace, namespace, EnvNacosGroup, group,
			EnvNacosDataID, dataID, "nacosServers", nacosServers)
	}
	client := &NacosClient{
		client:         nil,
		servers:        nacosServers,
		namespace:      namespace,
		group:          group,
		dataID:         dataID,
		log:            log,
		changeListener: p.ChangeListener,
		logLevel:       p.LogLevel,
		nacosLogger:    p.NacosLogger,
	}
	err := client.createNacosConfigClient()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (n *NacosClient) readConfig() (string, error) {
	if n.changeListener != nil {
		n.log.Infow("begin setup nacos config change listener")
		err := n.client.ListenConfig(vo.ConfigParam{
			DataId:   n.dataID,
			Group:    n.group,
			OnChange: n.changeListener,
		})
		if err != nil {
			return "", fmt.Errorf("nacos ListenConfig failed, err=%w", err)
		}
	}

	n.log.Infow("begin get config via nacos api")
	return n.client.GetConfig(vo.ConfigParam{
		DataId: n.dataID,
		Group:  n.group,
	})
}

func (n *NacosClient) createNacosConfigClient() error {
	clientConfig := constant.ClientConfig{
		NamespaceId:         n.namespace,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            n.logLevel,
		CustomLogger:        n.nacosLogger,
	}

	serverConfigs := make([]constant.ServerConfig, len(n.servers))
	for i, s := range n.servers {
		strs := strings.Split(s, ":")
		addr := strs[0]
		var port uint64 = 80
		if len(strs) > 1 {
			var err error
			port, err = strconv.ParseUint(strs[1], 10, 64)
			if err != nil {
				n.log.Errorw("failed to parse nacos server port %s, got error (%s), using default port 80", strs[1], err)
			}
		}

		serverConfigs[i] = constant.ServerConfig{
			IpAddr:      addr,
			Port:        port,
			ContextPath: "/nacos",
		}
	}

	nc, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	n.client = nc
	return err
}
