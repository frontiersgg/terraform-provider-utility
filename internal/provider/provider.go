// Copyright (c) Frontiers.gg
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func New(v string) func() provider.Provider {
	return func() provider.Provider {
		return &fileDownloaderProvider{
			version: v,
		}
	}
}

var _ provider.Provider = (*fileDownloaderProvider)(nil)

type fileDownloaderProvider struct {
	version string
}

func (p *fileDownloaderProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "utility"
	resp.Version = p.version
}

func (p *fileDownloaderProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
The Utility provider offers various utility functions and tools for use in Terraform configurations. This provider does not require configuration.
`,
	}
}

func (p *fileDownloaderProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
}

func (p *fileDownloaderProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewFileDownloaderResource,
	}
}

func (p *fileDownloaderProvider) DataSources(context.Context) []func() datasource.DataSource {
	return nil
}

type fileChecksums struct {
	sha1Hex   string
	sha256Hex string
}

func genFileChecksums(data []byte) *fileChecksums {
	sha1Sum := sha1.Sum(data)
	sha256Sum := sha256.Sum256(data)

	return &fileChecksums{
		sha1Hex:   hex.EncodeToString(sha1Sum[:]),
		sha256Hex: hex.EncodeToString(sha256Sum[:]),
	}
}
