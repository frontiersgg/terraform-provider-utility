// Copyright (c) Frontiers.gg
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type fileDownloaderResource struct{}

func NewFileDownloaderResource() resource.Resource {
	return &fileDownloaderResource{}
}

func (r *fileDownloaderResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "utility_file_downloader"
}

func (r *fileDownloaderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Resource to download a remote file via HTTP(S) using GET or POST, optionally with custom headers.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "The full HTTP or HTTPS URL to download the file from.",
				Required:    true,
			},
			"filename": schema.StringAttribute{
				Description: "Local filename where the downloaded file will be saved.",
				Required:    true,
			},
			"method": schema.StringAttribute{
				Description: "HTTP method to use for the request (default: GET). Only 'GET' and 'POST' are allowed.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(http.MethodGet, http.MethodPost),
				},
				Default: stringdefault.StaticString(http.MethodGet),
			},
			"headers": schema.MapAttribute{
				Description: "Map of custom HTTP headers to include in the request. The map key is the header name, and the value is the header content.",
				Optional:    true,
				ElementType: types.StringType,
				Sensitive:   true,
			},
			"force_download": schema.BoolAttribute{
				Description: "Force download even if the file url has not changed.",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Description: "The hexadecimal encoding of the SHA1 checksum of the downloaded file content.",
				Computed:    true,
			},
			"sha1": schema.StringAttribute{
				Description: "SHA1 checksum of file content.",
				Computed:    true,
			},
			"sha256": schema.StringAttribute{
				Description: "SHA256 checksum of file content.",
				Computed:    true,
			},
		},
	}
}

func (r *fileDownloaderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan fileResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	method := "GET"
	if !plan.Method.IsNull() && plan.Method.ValueString() != "" {
		method = strings.ToUpper(plan.Method.ValueString())
	}

	headers := make(map[string]string)
	for k, v := range plan.Headers.Elements() {
		if strVal, ok := v.(types.String); ok {
			headers[k] = strVal.ValueString()
		}
	}

	checksums, err := downloadFile(method, plan.URL.ValueString(), plan.Filename.ValueString(), headers)
	if err != nil {
		resp.Diagnostics.AddError("Download Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(checksums.sha1Hex)
	plan.Sha1 = types.StringValue(checksums.sha1Hex)
	plan.Sha256 = types.StringValue(checksums.sha256Hex)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *fileDownloaderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state fileResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	outputPath := state.Filename.ValueString()
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	method := "GET"
	if !state.Method.IsNull() && state.Method.ValueString() != "" {
		method = strings.ToUpper(state.Method.ValueString())
	}

	headers := make(map[string]string)
	for k, v := range state.Headers.Elements() {
		if strVal, ok := v.(types.String); ok {
			headers[k] = strVal.ValueString()
		}
	}

	checksums, err := downloadFile(method, state.URL.ValueString(), state.Filename.ValueString(), headers)
	if err != nil {
		resp.Diagnostics.AddError("Download Failed", err.Error())
		return
	}

	if checksums.sha1Hex != state.ID.ValueString() {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(checksums.sha1Hex)
	state.Sha1 = types.StringValue(checksums.sha1Hex)
	state.Sha256 = types.StringValue(checksums.sha256Hex)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *fileDownloaderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan fileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var state fileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	method := "GET"
	if !plan.Method.IsNull() && plan.Method.ValueString() != "" {
		method = strings.ToUpper(plan.Method.ValueString())
	}

	headers := make(map[string]string)
	for k, v := range plan.Headers.Elements() {
		if strVal, ok := v.(types.String); ok {
			headers[k] = strVal.ValueString()
		}
	}

	if !state.ForceDownload.ValueBool() && plan.URL.ValueString() == state.URL.ValueString() {
		resp.Diagnostics.AddWarning("same file", plan.URL.ValueString())
		resp.State.Set(ctx, state)
		return
	}

	checksums, err := downloadFile(method, plan.URL.ValueString(), plan.Filename.ValueString(), headers)
	if err != nil {
		resp.Diagnostics.AddError("Download Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(checksums.sha1Hex)
	plan.Sha1 = types.StringValue(checksums.sha1Hex)
	plan.Sha256 = types.StringValue(checksums.sha256Hex)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *fileDownloaderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var filename string
	req.State.GetAttribute(ctx, path.Root("filename"), &filename)
	os.Remove(filename)
}

type fileResourceModel struct {
	URL           types.String `tfsdk:"url"`
	Filename      types.String `tfsdk:"filename"`
	Method        types.String `tfsdk:"method"`
	Headers       types.Map    `tfsdk:"headers"`
	ForceDownload types.Bool   `tfsdk:"force_download"`
	ID            types.String `tfsdk:"id"`
	Sha1          types.String `tfsdk:"sha1"`
	Sha256        types.String `tfsdk:"sha256"`
}

func downloadFile(method, url, path string, headers map[string]string) (*fileChecksums, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to download file: " + resp.Status)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	out, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	checksums := genFileChecksums(bs)
	_, err = out.Write(bs)

	return checksums, err
}
