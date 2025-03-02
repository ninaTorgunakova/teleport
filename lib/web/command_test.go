/*
 * Copyright 2023 Gravitational, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package web

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/gravitational/roundtrip"
	"github.com/gravitational/trace"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gravitational/teleport/api/gen/proto/go/assist/v1"
	assistlib "github.com/gravitational/teleport/lib/assist"
	"github.com/gravitational/teleport/lib/auth"
	"github.com/gravitational/teleport/lib/client"
	"github.com/gravitational/teleport/lib/session"
	"github.com/gravitational/teleport/lib/utils"
)

func TestExecuteCommand(t *testing.T) {
	t.Parallel()
	s := newWebSuiteWithConfig(t, webSuiteConfig{disableDiskBasedRecording: true})

	ws, _, err := s.makeCommand(t, s.authPack(t, "foo"), uuid.New())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, ws.Close()) })

	stream := NewWStream(context.Background(), ws, utils.NewLoggerForTests(), nil)

	require.NoError(t, waitForCommandOutput(stream, "teleport"))
}

func TestExecuteCommandHistory(t *testing.T) {
	t.Parallel()

	// Given
	s := newWebSuiteWithConfig(t, webSuiteConfig{disableDiskBasedRecording: true})
	authPack := s.authPack(t, "foo")

	ctx := context.Background()
	clt, err := s.server.NewClient(auth.TestUser("foo"))
	require.NoError(t, err)

	// Create conversation, otherwise the command execution will not be saved
	conversation, err := clt.CreateAssistantConversation(context.Background(), &assist.CreateAssistantConversationRequest{
		Username:    "foo",
		CreatedTime: timestamppb.Now(),
	})
	require.NoError(t, err)

	require.NotEmpty(t, conversation.GetId())

	conversationID, err := uuid.Parse(conversation.GetId())
	require.NoError(t, err)

	ws, _, err := s.makeCommand(t, authPack, conversationID)
	require.NoError(t, err)

	stream := NewWStream(ctx, ws, utils.NewLoggerForTests(), nil)

	// When command executes
	require.NoError(t, waitForCommandOutput(stream, "teleport"))

	// Explecitly close the stream
	err = stream.Close()
	require.NoError(t, err)

	// Then command execution history is saved
	var messages *assist.GetAssistantMessagesResponse
	// Command execution history is saved in asynchronusly, so we need to wait for it.
	require.Eventually(t, func() bool {
		messages, err = clt.GetAssistantMessages(ctx, &assist.GetAssistantMessagesRequest{
			ConversationId: conversationID.String(),
			Username:       "foo",
		})
		require.NoError(t, err)

		return len(messages.GetMessages()) == 1
	}, 5*time.Second, 100*time.Millisecond)

	// Assert the returned message
	msg := messages.GetMessages()[0]
	require.Equal(t, string(assistlib.MessageKindCommandResult), msg.Type)
	require.NotZero(t, msg.CreatedTime)

	var result commandExecResult
	err = json.Unmarshal([]byte(msg.GetPayload()), &result)
	require.NoError(t, err)

	require.NotEmpty(t, result.ExecutionID)
	require.NotEmpty(t, result.SessionID)
	require.Equal(t, "node", result.NodeName)
	require.Equal(t, "node", result.NodeID)
}

func (s *WebSuite) makeCommand(t *testing.T, pack *authPack, conversationID uuid.UUID) (*websocket.Conn, *session.Session, error) {
	req := CommandRequest{
		Query:          fmt.Sprintf("name == \"%s\"", s.srvID),
		Login:          pack.login,
		ConversationID: conversationID.String(),
		ExecutionID:    uuid.New().String(),
		Command:        "echo txlxport | sed 's/x/e/g'",
	}

	u := url.URL{
		Host:   s.url().Host,
		Scheme: client.WSS,
		Path:   fmt.Sprintf("/v1/webapi/command/%v/execute", currentSiteShortcut),
	}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, nil, err
	}

	q := u.Query()
	q.Set("params", string(data))
	q.Set(roundtrip.AccessTokenQueryParam, pack.session.Token)
	u.RawQuery = q.Encode()

	dialer := websocket.Dialer{}
	dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	header := http.Header{}
	header.Add("Origin", "http://localhost")
	for _, cookie := range pack.cookies {
		header.Add("Cookie", cookie.String())
	}

	ws, resp, err := dialer.Dial(u.String(), header)
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}

	ty, raw, err := ws.ReadMessage()
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}
	require.Equal(t, websocket.BinaryMessage, ty)
	var env Envelope

	err = proto.Unmarshal(raw, &env)
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}

	var sessResp siteSessionGenerateResponse

	err = json.Unmarshal([]byte(env.Payload), &sessResp)
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}

	return ws, &sessResp.Session, nil
}

func waitForCommandOutput(stream io.Reader, substr string) error {
	timeoutCh := time.After(10 * time.Second)

	for {
		select {
		case <-timeoutCh:
			return trace.BadParameter("timeout waiting on terminal for output: %v", substr)
		default:
		}

		out := make([]byte, 100)
		n, err := stream.Read(out)
		if err != nil {
			return trace.Wrap(err)
		}

		var env Envelope
		err = json.Unmarshal(out[:n], &env)
		if err != nil {
			return trace.Wrap(err)
		}

		d, err := base64.StdEncoding.DecodeString(env.Payload)
		if err != nil {
			return trace.Wrap(err)
		}
		data := removeSpace(string(d))
		if n > 0 && strings.Contains(data, substr) {
			return nil
		}
		if err != nil {
			return trace.Wrap(err)
		}
	}
}

// Test_runCommands tests that runCommands runs the given command on all hosts.
// The commands should run in parallel, but we don't have a deterministic way to
// test that (sleep with checking the execution time in not deterministic).
func Test_runCommands(t *testing.T) {
	counter := atomic.Int32{}

	runCmd := func(host *hostInfo) error {
		counter.Add(1)
		return nil
	}

	hosts := make([]hostInfo, 0, 100)
	for i := 0; i < 100; i++ {
		hosts = append(hosts, hostInfo{
			hostName: fmt.Sprintf("localhost%d", i),
		})
	}

	logger := logrus.New()
	logger.Out = io.Discard

	runCommands(hosts, runCmd, logger)

	require.Equal(t, int32(100), counter.Load())
}
