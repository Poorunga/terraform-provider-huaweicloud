package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type DeleteDomainMappingRequest struct {

	// 直播播放域名
	PullDomain string `json:"pull_domain"`

	// 直播推流域名
	PushDomain string `json:"push_domain"`
}

func (o DeleteDomainMappingRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteDomainMappingRequest struct{}"
	}

	return strings.Join([]string{"DeleteDomainMappingRequest", string(data)}, " ")
}
