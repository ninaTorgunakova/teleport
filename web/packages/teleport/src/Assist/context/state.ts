/**
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

import { ServerMessageType } from 'teleport/Assist/types';

import type {
  Conversation,
  ResolvedAssistServerMessage,
  ResolvedCommandServerMessage,
  ResolvedServerMessage,
} from 'teleport/Assist/types';

export interface AssistState {
  conversations: {
    selectedId: string | null;
    error?: string;
    loading: boolean;
    data: Conversation[];
  };
  messages: {
    loading: boolean;
    error?: string;
    streaming: boolean;
    data: Map<string, ResolvedServerMessage[]>; // conversationId -> messages
  };
  mfa: {
    prompt: boolean;
    publicKey: PublicKeyCredentialRequestOptions | null;
  };
}

export enum AssistStateActionType {
  SetConversationsLoading,
  ReplaceConversations,
  SetSelectedConversationId,
  SetConversationMessagesLoading,
  SetConversationMessages,
  AddMessage,
  SetStreaming,
  AddPartialMessage,
  AddExecuteRemoteCommand,
  AddConversation,
  AddThought,
  AddCommandResult,
  UpdateCommandResult,
  FinishCommandResult,
  PromptMfa,
  DeleteConversation,
  UpdateConversationTitle,
}

export interface ReplaceConversationsAction {
  type: AssistStateActionType.ReplaceConversations;
  conversations: Conversation[];
}

export interface SetSelectedConversationIdAction {
  type: AssistStateActionType.SetSelectedConversationId;
  conversationId: string;
}

export interface SetConversationsLoadingAction {
  type: AssistStateActionType.SetConversationsLoading;
  loading: boolean;
}

export interface SetConversationMessagesLoadingAction {
  type: AssistStateActionType.SetConversationMessagesLoading;
  loading: boolean;
}

export interface SetConversationMessagesAction {
  type: AssistStateActionType.SetConversationMessages;
  messages: ResolvedServerMessage[];
  conversationId: string;
}

export interface AddMessageAction {
  type: AssistStateActionType.AddMessage;
  messageType:
    | ServerMessageType.User
    | ServerMessageType.Assist
    | ServerMessageType.Error;
  message: string;
  conversationId: string;
}

export interface SetStreamingAction {
  type: AssistStateActionType.SetStreaming;
  streaming: boolean;
}

export interface AddPartialMessageAction {
  type: AssistStateActionType.AddPartialMessage;
  message: string;
  conversationId: string;
}

export interface AddExecuteRemoteCommandAction {
  type: AssistStateActionType.AddExecuteRemoteCommand;
  message: ResolvedCommandServerMessage;
  conversationId: string;
}

export interface AddConversationAction {
  type: AssistStateActionType.AddConversation;
  conversationId: string;
}

export interface AddThoughtAction {
  type: AssistStateActionType.AddThought;
  message: string;
  conversationId: string;
}

export interface AddCommandResultAction {
  type: AssistStateActionType.AddCommandResult;
  id: number;
  nodeId: string;
  nodeName: string;
  conversationId: string;
}

export interface UpdateCommandResultAction {
  type: AssistStateActionType.UpdateCommandResult;
  commandResultId: number;
  output: string;
  conversationId: string;
}

export interface FinishCommandResultAction {
  type: AssistStateActionType.FinishCommandResult;
  commandResultId: number;
  conversationId: string;
}

export interface PromptMfaAction {
  type: AssistStateActionType.PromptMfa;
  publicKey: PublicKeyCredentialRequestOptions | null;
  promptMfa: boolean;
}

export interface DeleteConversationAction {
  type: AssistStateActionType.DeleteConversation;
  conversationId: string;
}

export interface UpdateConversationTitleAction {
  type: AssistStateActionType.UpdateConversationTitle;
  conversationId: string;
  title: string;
}

export type AssistContextAction =
  | SetConversationsLoadingAction
  | ReplaceConversationsAction
  | SetSelectedConversationIdAction
  | SetConversationMessagesLoadingAction
  | SetConversationMessagesAction
  | AddMessageAction
  | SetStreamingAction
  | AddPartialMessageAction
  | AddExecuteRemoteCommandAction
  | AddConversationAction
  | AddThoughtAction
  | AddCommandResultAction
  | UpdateCommandResultAction
  | FinishCommandResultAction
  | PromptMfaAction
  | DeleteConversationAction
  | UpdateConversationTitleAction;

export function reducer(
  state: AssistState,
  action: AssistContextAction
): AssistState {
  switch (action.type) {
    case AssistStateActionType.SetConversationsLoading:
      return setConversationsLoading(state, action);

    case AssistStateActionType.ReplaceConversations:
      return replaceConversations(state, action);

    case AssistStateActionType.SetSelectedConversationId:
      return setSelectedConversationId(state, action);

    case AssistStateActionType.SetConversationMessagesLoading:
      return setConversationMessagesLoading(state, action);

    case AssistStateActionType.SetConversationMessages:
      return setConversationMessages(state, action);

    case AssistStateActionType.AddMessage:
      return addMessage(state, action);

    case AssistStateActionType.SetStreaming:
      return setStreaming(state, action);

    case AssistStateActionType.AddPartialMessage:
      return addPartialMessage(state, action);

    case AssistStateActionType.AddExecuteRemoteCommand:
      return addExecuteRemoteCommand(state, action);

    case AssistStateActionType.AddConversation:
      return addConversation(state, action);

    case AssistStateActionType.AddThought:
      return addThought(state, action);

    case AssistStateActionType.AddCommandResult:
      return addCommandResult(state, action);

    case AssistStateActionType.UpdateCommandResult:
      return updateCommandResult(state, action);

    case AssistStateActionType.FinishCommandResult:
      return finishCommandResult(state, action);

    case AssistStateActionType.PromptMfa:
      return promptMfa(state, action);

    case AssistStateActionType.DeleteConversation:
      return deleteConversation(state, action);

    case AssistStateActionType.UpdateConversationTitle:
      return updateConversationTitle(state, action);

    default:
      return state;
  }
}

export function setConversationsLoading(
  state: AssistState,
  action: SetConversationsLoadingAction
): AssistState {
  return {
    ...state,
    conversations: {
      ...state.conversations,
      loading: action.loading,
    },
  };
}

export function replaceConversations(
  state: AssistState,
  action: ReplaceConversationsAction
): AssistState {
  return {
    ...state,
    conversations: {
      selectedId: state.conversations.selectedId,
      loading: false,
      error: null,
      data: action.conversations,
    },
  };
}

export function setSelectedConversationId(
  state: AssistState,
  action: SetSelectedConversationIdAction
): AssistState {
  return {
    ...state,
    conversations: {
      ...state.conversations,
      selectedId: action.conversationId,
    },
  };
}

export function setConversationMessagesLoading(
  state: AssistState,
  action: SetConversationMessagesLoadingAction
): AssistState {
  return {
    ...state,
    messages: {
      ...state.messages,
      loading: action.loading,
    },
  };
}

export function setConversationMessages(
  state: AssistState,
  action: SetConversationMessagesAction
): AssistState {
  const messages = new Map(state.messages.data);

  messages.set(action.conversationId, action.messages);

  return {
    ...state,
    messages: {
      loading: false,
      streaming: false,
      data: messages,
    },
  };
}

export function addMessage(
  state: AssistState,
  action: AddMessageAction
): AssistState {
  const messages = new Map(state.messages.data);

  if (messages.has(action.conversationId)) {
    const existingMessages = messages.get(action.conversationId);

    messages.set(action.conversationId, [
      ...existingMessages,
      {
        type: action.messageType,
        message: action.message,
        created: new Date(),
      },
    ]);
  } else {
    messages.set(action.conversationId, [
      {
        type: action.messageType,
        message: action.message,
        created: new Date(),
      },
    ]);
  }

  return {
    ...state,
    messages: {
      ...state.messages,
      data: messages,
    },
  };
}

export function setStreaming(
  state: AssistState,
  action: SetStreamingAction
): AssistState {
  return {
    ...state,
    messages: {
      ...state.messages,
      streaming: action.streaming,
    },
  };
}

export function addPartialMessage(
  state: AssistState,
  action: AddPartialMessageAction
): AssistState {
  const messages = new Map(state.messages.data);

  let conversationMessages = messages.get(action.conversationId);

  if (
    conversationMessages[conversationMessages.length - 1].type ===
    ServerMessageType.Assist
  ) {
    conversationMessages = conversationMessages.map(
      (message: ResolvedAssistServerMessage, index) => {
        if (index === conversationMessages.length - 1) {
          return {
            ...message,
            message: message.message + action.message,
          };
        }

        return message;
      }
    );
  } else {
    conversationMessages = [
      ...conversationMessages,
      {
        type: ServerMessageType.Assist,
        message: action.message,
        created: new Date(),
      },
    ];
  }

  messages.set(action.conversationId, conversationMessages);

  return {
    ...state,
    messages: {
      ...state.messages,
      data: messages,
    },
  };
}

export function addExecuteRemoteCommand(
  state: AssistState,
  action: AddExecuteRemoteCommandAction
): AssistState {
  const messages = new Map(state.messages.data);

  let conversationMessages = messages.get(action.conversationId);

  conversationMessages = [
    ...conversationMessages,
    {
      type: ServerMessageType.Command,
      created: new Date(),
      query: action.message.query,
      command: action.message.command,
    },
  ];

  messages.set(action.conversationId, conversationMessages);

  return {
    ...state,
    messages: {
      ...state.messages,
      data: messages,
    },
  };
}

export function addConversation(
  state: AssistState,
  action: AddConversationAction
): AssistState {
  const conversations = state.conversations.data;

  return {
    ...state,
    conversations: {
      ...state.conversations,
      selectedId: action.conversationId,
      data: [
        {
          id: action.conversationId,
          title: 'New conversation',
          created: new Date(),
        },
        ...conversations,
      ],
    },
  };
}

export function addThought(
  state: AssistState,
  action: AddThoughtAction
): AssistState {
  const messages = new Map(state.messages.data);

  let conversationMessages = messages.get(action.conversationId);

  conversationMessages = [
    ...conversationMessages,
    {
      type: ServerMessageType.AssistThought,
      created: new Date(),
      message: action.message,
    },
  ];

  messages.set(action.conversationId, conversationMessages);

  return {
    ...state,
    messages: {
      ...state.messages,
      data: messages,
    },
  };
}

export function addCommandResult(
  state: AssistState,
  action: AddCommandResultAction
): AssistState {
  const messages = new Map(state.messages.data);

  let conversationMessages = messages.get(action.conversationId);

  conversationMessages = [
    ...conversationMessages,
    {
      type: ServerMessageType.CommandResultStream,
      id: action.id,
      nodeId: action.nodeId,
      nodeName: action.nodeName,
      created: new Date(),
      finished: false,
      output: '',
    },
  ];

  messages.set(action.conversationId, conversationMessages);

  return {
    ...state,
    messages: {
      ...state.messages,
      data: messages,
    },
  };
}

export function updateCommandResult(
  state: AssistState,
  action: UpdateCommandResultAction
): AssistState {
  const messages = new Map(state.messages.data);

  const conversationMessages = messages
    .get(action.conversationId)
    .map(message => {
      if (
        message.type === ServerMessageType.CommandResultStream &&
        message.id === action.commandResultId
      ) {
        return {
          ...message,
          output: message.output + action.output,
        };
      }

      return message;
    });

  messages.set(action.conversationId, conversationMessages);

  return {
    ...state,
    messages: {
      ...state.messages,
      data: messages,
    },
  };
}

export function finishCommandResult(
  state: AssistState,
  action: FinishCommandResultAction
): AssistState {
  const messages = new Map(state.messages.data);

  const conversationMessages = messages
    .get(action.conversationId)
    .map(message => {
      if (
        message.type === ServerMessageType.CommandResultStream &&
        message.id === action.commandResultId
      ) {
        return {
          ...message,
          finished: true,
        };
      }

      return message;
    });

  messages.set(action.conversationId, conversationMessages);

  return {
    ...state,
    messages: {
      ...state.messages,
      data: messages,
    },
  };
}

export function promptMfa(
  state: AssistState,
  action: PromptMfaAction
): AssistState {
  return {
    ...state,
    mfa: {
      prompt: action.promptMfa,
      publicKey: action.publicKey,
    },
  };
}

export function deleteConversation(
  state: AssistState,
  action: DeleteConversationAction
) {
  const conversations = state.conversations.data;

  const newSelectedId =
    state.conversations.selectedId === action.conversationId
      ? null
      : state.conversations.selectedId;

  return {
    ...state,
    conversations: {
      ...state.conversations,
      selectedId: newSelectedId,
      data: conversations.filter(
        conversation => conversation.id !== action.conversationId
      ),
    },
  };
}

export function updateConversationTitle(
  state: AssistState,
  action: UpdateConversationTitleAction
) {
  const conversations = state.conversations.data;

  return {
    ...state,
    conversations: {
      ...state.conversations,
      data: conversations.map(conversation => {
        if (conversation.id === action.conversationId) {
          return {
            ...conversation,
            title: action.title,
          };
        }

        return conversation;
      }),
    },
  };
}
