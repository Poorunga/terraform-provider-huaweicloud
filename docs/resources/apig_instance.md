---
subcategory: "API Gateway (Dedicated APIG)"
---

# huaweicloud_apig_instance

Manages an APIG dedicated instance resource within HuaweiCloud.

## Example Usage

```hcl
variable "instance_name" {}
variable "vpc_id" {}
variable "subnet_id" {}
variable "security_group_id" {}
variable "eip_id" {}
variable "enterprise_project_id" {}

data "huaweicloud_availability_zones" "test" {}

resource "huaweicloud_apig_instance" "test" {
  name                  = var.instance_name
  edition               = "BASIC"
  vpc_id                = var.vpc_id
  subnet_id             = var.subnet_id
  security_group_id     = var.security_group_id
  enterprise_project_id = var.enterprise_project_id
  maintain_begin        = "06:00:00"
  description           = "Created by script"
  bandwidth_size        = 3
  eip_id                = var.eip_id

  available_zones = [
    data.huaweicloud_availability_zones.test.names[0],
    data.huaweicloud_availability_zones.test.names[1],
  ]
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional, String, ForceNew) Specifies the region in which to create the dedicated instance resource.  
  If omitted, the provider-level region will be used.
  Changing this will create a new resource.

* `name` - (Required, String) Specifies the name of the dedicated instance.  
  The name can contain `3` to `64` characters, only letters, digits, hyphens (-) and underscores (_) are allowed, and
  must start with a letter.

* `edition` - (Required, String, ForceNew) Specifies the edition of the dedicated instance.  
  The valid values are as follows:
  + **BASIC**: Basic Edition instance.
  + **PROFESSIONAL**: Professional Edition instance.
  + **ENTERPRISE**: Enterprise Edition instance.
  + **PLATINUM**: Platinum Edition instance.
  + **BASIC_IPV6**: IPv6 instance of the Basic Edition.
  + **PROFESSIONAL_IPV6**: IPv6 instance of the Professional Edition.
  + **ENTERPRISE_IPV6**: IPv6 instance of the Enterprise Edition.
  + **PLATINUM_IPV6**: IPv6 instance of the Platinum Edition.
  
  Changing this will create a new resource.

* `vpc_id` - (Required, String, ForceNew) Specifies the ID of the VPC used to create the dedicated instance.  
  Changing this will create a new resource.

* `subnet_id` - (Required, String, ForceNew) Specifies the ID of the VPC subnet used to create the dedicated instance.  
  Changing this will create a new resource.

* `security_group_id` - (Required, String) Specifies the ID of the security group to which the dedicated instance
  belongs to.

* `availability_zones` - (Required, List, ForceNew) Specifies the name list of availability zones for the dedicated
  instance.  
  Please following [reference](https://developer.huaweicloud.com/intl/en-us/endpoint?APIG) for list elements.
  Changing this will create a new resource.

* `description` - (Optional, String) Specifies the description of the dedicated instance.  
  The description contain a maximum of `255` characters and the angle brackets (< and >) are not allowed.

* `enterprise_project_id` - (Optional, String, ForceNew) Specifies the enterprise project ID to which the dedicated
  instance belongs.  
  This parameter is required for enterprise users. Changing this will create a new resource.

* `bandwidth_size` - (Optional, Int) Specifies the egress bandwidth size of the dedicated instance.  
  The valid value is range from `1` to `2,000`.

* `maintain_begin` - (Optional, String) Specifies the start time of the maintenance time window.  
  The format is **xx:00:00**, the value of **xx** can be `02`, `06`, `10`, `14`, `18` or `22`.

* `eip_id` - (Optional, String) Specifies the EIP ID associated with the dedicated instance.

* `ipv6_enable` - (Optional, Bool, ForceNew) Specifies whether public access with an IPv6 address is supported.  
  Changing this will create a new resource.

* `loadbalancer_provider` - (Optional, String, ForceNew) Specifies the provider type of load balancer used by the
  dedicated instance.  
  The valid values are as follows:
  + **lvs**: Linux virtual server.
  + **elb**: Elastic load balance.

  Changing this will create a new resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the dedicated instance.
* `maintain_end` - End time of the maintenance time window, 4-hour difference between the start time and end time.
* `ingress_address` - The ingress EIP address.
* `vpc_ingress_address` - The ingress private IP address of the VPC.
* `egress_address` - The egress (NAT) public IP address.
* `supported_features` - The supported features of the APIG dedicated instance.
* `created_at` - Time when the dedicated instance is created, in RFC-3339 format.
* `status` - Status of the dedicated instance.

## Timeouts

This resource provides the following timeouts configuration options:

* `create` - Default is 40 minutes.
* `update` - Default is 10 minutes.
* `delete` - Default is 10 minutes.

## Import

Dedicated instances can be imported by their `id`, e.g.

```
$ terraform import huaweicloud_apig_instance.test de379eed30aa4d31a84f426ea3c7ef4e
```
