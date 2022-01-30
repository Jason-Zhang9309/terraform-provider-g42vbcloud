---
subcategory: "Virtual Private Cloud (VPC)"
---

# g42vbcloud_networking_secgroup_rule

Manages a Security Group Rule resource within within G42VBCloud.

## Example Usage

```hcl
resource "g42vbcloud_networking_secgroup" "secgroup_1" {
  name        = "secgroup_1"
  description = "My neutron security group"
}

resource "g42vbcloud_networking_secgroup_rule" "secgroup_rule_1" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = g42vbcloud_networking_secgroup.secgroup_1.id
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional, String, ForceNew) Specifies the region in which to create the security group rule resource. If
  omitted, the provider-level region will be used. Changing this creates a new security group rule.

* `security_group_id` - (Required, String, ForceNew) Specifies the security group id the rule should belong to. Changing
  this creates a new security group rule.

* `direction` - (Required, String, ForceNew) Specifies the direction of the rule, valid values are __ingress__ or
  __egress__. Changing this creates a new security group rule.

* `ethertype` - (Required, String, ForceNew) Specifies the layer 3 protocol type, valid values are __IPv4__ or __IPv6__.
  Changing this creates a new security group rule.

* `description` - (Optional, String, ForceNew) Specifies the supplementary information about the networking security
  group rule. This parameter can contain a maximum of 255 characters and cannot contain angle brackets (< or >).
  Changing this creates a new security group rule.

* `protocol` - (Optional, String, ForceNew) Specifies the layer 4 protocol type, valid values are __tcp__, __udp__,
  __icmp__ and __icmpv6__. If omitted, the protocol means that all protocols are supported.
  This is required if you want to specify a port range. Changing this creates a new security group rule.

* `port_range_min` - (Optional, Int, ForceNew) Specifies the lower part of the allowed port range, valid integer value
  needs to be between 1 and 65535. Changing this creates a new security group rule.

* `port_range_max` - (Optional, Int, ForceNew) Specifies the higher part of the allowed port range, valid integer value
  needs to be between 1 and 65535. Changing this creates a new security group rule.

* `remote_ip_prefix` - (Optional, String, ForceNew) Specifies the remote CIDR, the value needs to be a valid CIDR (i.e.
  192.168.0.0/16). Changing this creates a new security group rule.

* `remote_group_id` - (Optional, String, ForceNew) Specifies the remote group id. Changing this creates a new security
  group rule.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource ID in UUID format.

## Timeouts

This resource provides the following timeouts configuration options:

* `delete` - Default is 10 minute.

## Import

Security Group Rules can be imported using the `id`, e.g.

```
$ terraform import g42vbcloud_networking_secgroup_rule.secgroup_rule_1 aeb68ee3-6e9d-4256-955c-9584a6212745
```
