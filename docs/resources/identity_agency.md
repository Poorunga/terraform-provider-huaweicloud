---
subcategory: "Identity and Access Management (IAM)"
---

# huaweicloud_identity_agency

Manages an agency resource within huawei cloud.

## Example Usage

### Delegate another HUAWEI CLOUD account to perform operations on your resources

```hcl
resource "huaweicloud_identity_agency" "agency" {
  name                  = "test_agency"
  description           = "test agency"
  delegated_domain_name = "***"

  project_role {
    project = "cn-north-1"
    roles   = ["Tenant Administrator"]
  }
  domain_roles = ["Anti-DDoS Administrator"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String, ForceNew) Specifies the name of agency. The name is a string of 1 to 64 characters.
  Changing this will create a new agency.

* `description` - (Optional, String) Specifies the supplementary information about the agency. The value is a string of
  0 to 255 characters, excluding these characters: '**@#$%^&*<>\\**'.

* `delegated_domain_name` - (Required, String) Specifies the name of delegated user domain.

* `duration` - (Optional, String) Specifies the validity period of an agency. The valid value are *ONEDAY* and *FOREVER*
  , defaults to *FOREVER*.

* `project_role` - (Optional, List) Specifies an array of one or more roles and projects which are used to grant
  permissions to agency on project. The structure is documented below.

* `domain_roles` - (Optional, List) Specifies an array of one or more role names which stand for the permissions to be
  granted to agency on domain.

The `project_role` block supports:

* `project` - (Required, String) Specifies the name of project.

* `roles` - (Required, List) Specifies an array of role names.

-> **NOTE**
At least one of `project_role` and `domain_roles` must be specified when creating an agency. We can get all **
System-Defined Roles** form
[HuaweiCloud](https://support.huaweicloud.com/intl/en-us/usermanual-permissions/iam_01_0001.html).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The agency ID.
* `expire_time` - The expiration time of agency.
* `create_time` - The time when the agency was created.

## Import

Agencies can be imported using the `id`, e.g.

```
$ terraform import huaweicloud_identity_agency.agency 0b97661f9900f23f4fc2c00971ea4dc0
```

## Timeouts

This resource provides the following timeouts configuration options:

* `create` - Default is 10 minute.
* `update` - Default is 10 minute.
* `delete` - Default is 5 minute.
