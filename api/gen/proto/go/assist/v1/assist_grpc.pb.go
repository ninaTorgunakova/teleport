// Copyright 2023 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: teleport/assist/v1/assist.proto

package assist

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	AssistService_CreateAssistantConversation_FullMethodName     = "/teleport.assist.v1.AssistService/CreateAssistantConversation"
	AssistService_GetAssistantConversations_FullMethodName       = "/teleport.assist.v1.AssistService/GetAssistantConversations"
	AssistService_DeleteAssistantConversation_FullMethodName     = "/teleport.assist.v1.AssistService/DeleteAssistantConversation"
	AssistService_GetAssistantMessages_FullMethodName            = "/teleport.assist.v1.AssistService/GetAssistantMessages"
	AssistService_CreateAssistantMessage_FullMethodName          = "/teleport.assist.v1.AssistService/CreateAssistantMessage"
	AssistService_UpdateAssistantConversationInfo_FullMethodName = "/teleport.assist.v1.AssistService/UpdateAssistantConversationInfo"
	AssistService_IsAssistEnabled_FullMethodName                 = "/teleport.assist.v1.AssistService/IsAssistEnabled"
)

// AssistServiceClient is the client API for AssistService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AssistServiceClient interface {
	// CreateNewConversation creates a new conversation and returns the UUID of it.
	CreateAssistantConversation(ctx context.Context, in *CreateAssistantConversationRequest, opts ...grpc.CallOption) (*CreateAssistantConversationResponse, error)
	// GetAssistantConversations returns all conversations for the connected user.
	GetAssistantConversations(ctx context.Context, in *GetAssistantConversationsRequest, opts ...grpc.CallOption) (*GetAssistantConversationsResponse, error)
	// DeleteAssistantConversation deletes the conversation and all messages associated with it.
	DeleteAssistantConversation(ctx context.Context, in *DeleteAssistantConversationRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// GetAssistantMessages returns all messages associated with the given conversation ID.
	GetAssistantMessages(ctx context.Context, in *GetAssistantMessagesRequest, opts ...grpc.CallOption) (*GetAssistantMessagesResponse, error)
	// CreateAssistantMessage creates a new message in the given conversation.
	CreateAssistantMessage(ctx context.Context, in *CreateAssistantMessageRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// UpdateAssistantConversationInfo updates the conversation info.
	UpdateAssistantConversationInfo(ctx context.Context, in *UpdateAssistantConversationInfoRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// IsAssistEnabled returns true if the assist is enabled or not on the auth level.
	IsAssistEnabled(ctx context.Context, in *IsAssistEnabledRequest, opts ...grpc.CallOption) (*IsAssistEnabledResponse, error)
}

type assistServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAssistServiceClient(cc grpc.ClientConnInterface) AssistServiceClient {
	return &assistServiceClient{cc}
}

func (c *assistServiceClient) CreateAssistantConversation(ctx context.Context, in *CreateAssistantConversationRequest, opts ...grpc.CallOption) (*CreateAssistantConversationResponse, error) {
	out := new(CreateAssistantConversationResponse)
	err := c.cc.Invoke(ctx, AssistService_CreateAssistantConversation_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *assistServiceClient) GetAssistantConversations(ctx context.Context, in *GetAssistantConversationsRequest, opts ...grpc.CallOption) (*GetAssistantConversationsResponse, error) {
	out := new(GetAssistantConversationsResponse)
	err := c.cc.Invoke(ctx, AssistService_GetAssistantConversations_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *assistServiceClient) DeleteAssistantConversation(ctx context.Context, in *DeleteAssistantConversationRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, AssistService_DeleteAssistantConversation_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *assistServiceClient) GetAssistantMessages(ctx context.Context, in *GetAssistantMessagesRequest, opts ...grpc.CallOption) (*GetAssistantMessagesResponse, error) {
	out := new(GetAssistantMessagesResponse)
	err := c.cc.Invoke(ctx, AssistService_GetAssistantMessages_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *assistServiceClient) CreateAssistantMessage(ctx context.Context, in *CreateAssistantMessageRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, AssistService_CreateAssistantMessage_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *assistServiceClient) UpdateAssistantConversationInfo(ctx context.Context, in *UpdateAssistantConversationInfoRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, AssistService_UpdateAssistantConversationInfo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *assistServiceClient) IsAssistEnabled(ctx context.Context, in *IsAssistEnabledRequest, opts ...grpc.CallOption) (*IsAssistEnabledResponse, error) {
	out := new(IsAssistEnabledResponse)
	err := c.cc.Invoke(ctx, AssistService_IsAssistEnabled_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AssistServiceServer is the server API for AssistService service.
// All implementations must embed UnimplementedAssistServiceServer
// for forward compatibility
type AssistServiceServer interface {
	// CreateNewConversation creates a new conversation and returns the UUID of it.
	CreateAssistantConversation(context.Context, *CreateAssistantConversationRequest) (*CreateAssistantConversationResponse, error)
	// GetAssistantConversations returns all conversations for the connected user.
	GetAssistantConversations(context.Context, *GetAssistantConversationsRequest) (*GetAssistantConversationsResponse, error)
	// DeleteAssistantConversation deletes the conversation and all messages associated with it.
	DeleteAssistantConversation(context.Context, *DeleteAssistantConversationRequest) (*emptypb.Empty, error)
	// GetAssistantMessages returns all messages associated with the given conversation ID.
	GetAssistantMessages(context.Context, *GetAssistantMessagesRequest) (*GetAssistantMessagesResponse, error)
	// CreateAssistantMessage creates a new message in the given conversation.
	CreateAssistantMessage(context.Context, *CreateAssistantMessageRequest) (*emptypb.Empty, error)
	// UpdateAssistantConversationInfo updates the conversation info.
	UpdateAssistantConversationInfo(context.Context, *UpdateAssistantConversationInfoRequest) (*emptypb.Empty, error)
	// IsAssistEnabled returns true if the assist is enabled or not on the auth level.
	IsAssistEnabled(context.Context, *IsAssistEnabledRequest) (*IsAssistEnabledResponse, error)
	mustEmbedUnimplementedAssistServiceServer()
}

// UnimplementedAssistServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAssistServiceServer struct {
}

func (UnimplementedAssistServiceServer) CreateAssistantConversation(context.Context, *CreateAssistantConversationRequest) (*CreateAssistantConversationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAssistantConversation not implemented")
}
func (UnimplementedAssistServiceServer) GetAssistantConversations(context.Context, *GetAssistantConversationsRequest) (*GetAssistantConversationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAssistantConversations not implemented")
}
func (UnimplementedAssistServiceServer) DeleteAssistantConversation(context.Context, *DeleteAssistantConversationRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteAssistantConversation not implemented")
}
func (UnimplementedAssistServiceServer) GetAssistantMessages(context.Context, *GetAssistantMessagesRequest) (*GetAssistantMessagesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAssistantMessages not implemented")
}
func (UnimplementedAssistServiceServer) CreateAssistantMessage(context.Context, *CreateAssistantMessageRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAssistantMessage not implemented")
}
func (UnimplementedAssistServiceServer) UpdateAssistantConversationInfo(context.Context, *UpdateAssistantConversationInfoRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateAssistantConversationInfo not implemented")
}
func (UnimplementedAssistServiceServer) IsAssistEnabled(context.Context, *IsAssistEnabledRequest) (*IsAssistEnabledResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsAssistEnabled not implemented")
}
func (UnimplementedAssistServiceServer) mustEmbedUnimplementedAssistServiceServer() {}

// UnsafeAssistServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AssistServiceServer will
// result in compilation errors.
type UnsafeAssistServiceServer interface {
	mustEmbedUnimplementedAssistServiceServer()
}

func RegisterAssistServiceServer(s grpc.ServiceRegistrar, srv AssistServiceServer) {
	s.RegisterService(&AssistService_ServiceDesc, srv)
}

func _AssistService_CreateAssistantConversation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateAssistantConversationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AssistServiceServer).CreateAssistantConversation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AssistService_CreateAssistantConversation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AssistServiceServer).CreateAssistantConversation(ctx, req.(*CreateAssistantConversationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AssistService_GetAssistantConversations_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAssistantConversationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AssistServiceServer).GetAssistantConversations(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AssistService_GetAssistantConversations_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AssistServiceServer).GetAssistantConversations(ctx, req.(*GetAssistantConversationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AssistService_DeleteAssistantConversation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteAssistantConversationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AssistServiceServer).DeleteAssistantConversation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AssistService_DeleteAssistantConversation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AssistServiceServer).DeleteAssistantConversation(ctx, req.(*DeleteAssistantConversationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AssistService_GetAssistantMessages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAssistantMessagesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AssistServiceServer).GetAssistantMessages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AssistService_GetAssistantMessages_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AssistServiceServer).GetAssistantMessages(ctx, req.(*GetAssistantMessagesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AssistService_CreateAssistantMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateAssistantMessageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AssistServiceServer).CreateAssistantMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AssistService_CreateAssistantMessage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AssistServiceServer).CreateAssistantMessage(ctx, req.(*CreateAssistantMessageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AssistService_UpdateAssistantConversationInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateAssistantConversationInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AssistServiceServer).UpdateAssistantConversationInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AssistService_UpdateAssistantConversationInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AssistServiceServer).UpdateAssistantConversationInfo(ctx, req.(*UpdateAssistantConversationInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AssistService_IsAssistEnabled_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IsAssistEnabledRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AssistServiceServer).IsAssistEnabled(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AssistService_IsAssistEnabled_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AssistServiceServer).IsAssistEnabled(ctx, req.(*IsAssistEnabledRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AssistService_ServiceDesc is the grpc.ServiceDesc for AssistService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AssistService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "teleport.assist.v1.AssistService",
	HandlerType: (*AssistServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateAssistantConversation",
			Handler:    _AssistService_CreateAssistantConversation_Handler,
		},
		{
			MethodName: "GetAssistantConversations",
			Handler:    _AssistService_GetAssistantConversations_Handler,
		},
		{
			MethodName: "DeleteAssistantConversation",
			Handler:    _AssistService_DeleteAssistantConversation_Handler,
		},
		{
			MethodName: "GetAssistantMessages",
			Handler:    _AssistService_GetAssistantMessages_Handler,
		},
		{
			MethodName: "CreateAssistantMessage",
			Handler:    _AssistService_CreateAssistantMessage_Handler,
		},
		{
			MethodName: "UpdateAssistantConversationInfo",
			Handler:    _AssistService_UpdateAssistantConversationInfo_Handler,
		},
		{
			MethodName: "IsAssistEnabled",
			Handler:    _AssistService_IsAssistEnabled_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "teleport/assist/v1/assist.proto",
}
