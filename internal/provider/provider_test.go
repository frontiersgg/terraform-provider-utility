// Copyright (c) Frontiers.gg
// SPDX-License-Identifier: MIT

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var protoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"utility": providerserver.NewProtocol6WithError(New("test")()),
}
