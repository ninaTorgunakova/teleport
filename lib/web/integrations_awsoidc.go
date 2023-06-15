/*
Copyright 2023 Gravitational, Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package web

import (
	"context"
	"net/http"

	"github.com/gravitational/trace"
	"github.com/julienschmidt/httprouter"

	"github.com/gravitational/teleport"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/httplib"
	"github.com/gravitational/teleport/lib/integrations/awsoidc"
	"github.com/gravitational/teleport/lib/reversetunnel"
	"github.com/gravitational/teleport/lib/web/ui"
)

// awsOIDCListDatabases returns a list of databases using the ListDatabases action of the AWS OIDC Integration.
func (h *Handler) awsOIDCListDatabases(w http.ResponseWriter, r *http.Request, p httprouter.Params, sctx *SessionContext, site reversetunnel.RemoteSite) (interface{}, error) {
	ctx := r.Context()

	var req ui.AWSOIDCListDatabasesRequest
	if err := httplib.ReadJSON(r, &req); err != nil {
		return nil, trace.Wrap(err)
	}

	awsClientReq, err := h.awsOIDCClientRequest(r.Context(), req.Region, p, sctx, site)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	listDBsClient, err := awsoidc.NewListDatabasesClient(ctx, awsClientReq)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	resp, err := awsoidc.ListDatabases(ctx,
		listDBsClient,
		awsoidc.ListDatabasesRequest{
			Region:    req.Region,
			NextToken: req.NextToken,
			Engines:   req.Engines,
			RDSType:   req.RDSType,
		},
	)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return ui.AWSOIDCListDatabasesResponse{
		NextToken: resp.NextToken,
		Databases: ui.MakeDatabases(resp.Databases, nil, nil),
	}, nil
}

// awsOIDClientRequest receives a request to execute an action for the AWS OIDC integrations.
func (h *Handler) awsOIDCClientRequest(ctx context.Context, region string, p httprouter.Params, sctx *SessionContext, site reversetunnel.RemoteSite) (*awsoidc.AWSClientRequest, error) {
	integrationName := p.ByName("name")
	if integrationName == "" {
		return nil, trace.BadParameter("an integration name is required")
	}

	clt, err := sctx.GetUserClient(ctx, site)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	integration, err := clt.GetIntegration(ctx, integrationName)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if integration.GetSubKind() != types.IntegrationSubKindAWSOIDC {
		return nil, trace.BadParameter("integration subkind (%s) mismatch", integration.GetSubKind())
	}

	issuer, err := h.issuerFromPublicAddr()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	token, err := clt.GenerateAWSOIDCToken(ctx, types.GenerateAWSOIDCTokenRequest{
		Issuer: issuer,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	awsoidcSpec := integration.GetAWSOIDCIntegrationSpec()
	if awsoidcSpec == nil {
		return nil, trace.BadParameter("missing spec fields for %q (%q) integration", integration.GetName(), integration.GetSubKind())
	}

	return &awsoidc.AWSClientRequest{
		Token:   token,
		RoleARN: awsoidcSpec.RoleARN,
		Region:  region,
	}, nil
}

// awsOIDCDeployDBService deploys a Discovery Service and a Database Service in Amazon ECS.
func (h *Handler) awsOIDCDeployDBService(w http.ResponseWriter, r *http.Request, p httprouter.Params, sctx *SessionContext, site reversetunnel.RemoteSite) (interface{}, error) {
	ctx := r.Context()

	var req ui.AWSOIDCDeployDBServiceRequest
	if err := httplib.ReadJSON(r, &req); err != nil {
		return nil, trace.Wrap(err)
	}

	awsClientReq, err := h.awsOIDCClientRequest(ctx, req.Region, p, sctx, site)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	deployDBServiceClient, err := awsoidc.NewDeployDBServiceClient(ctx, awsClientReq)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	deployDBServiceResp, err := awsoidc.DeployDBService(ctx, deployDBServiceClient, awsoidc.DeployDBServiceRequest{
		Region:              req.Region,
		SubnetIDs:           req.SubnetIDs,
		ClusterName:         req.ClusterName,
		ServiceName:         req.ServiceName,
		TaskName:            req.TaskName,
		TaskRoleARN:         req.TaskRoleARN,
		DiscoveryGroupName:  req.DiscoveryGroupName,
		ProxyServerHostPort: h.PublicProxyAddr(),
		TeleportVersion:     teleport.Version,
		TeleportClusterName: h.auth.clusterName,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return ui.AWSOIDCDeployDBServiceResponse{
		ClusterARN:        deployDBServiceResp.ClusterARN,
		ServiceARN:        deployDBServiceResp.ServiceARN,
		TaskDefinitionARN: deployDBServiceResp.TaskDefinitionARN,
	}, nil
}
