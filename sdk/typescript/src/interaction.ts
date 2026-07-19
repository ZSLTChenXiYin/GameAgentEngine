// Interaction mode constants
export const InteractionMode = {
  DirectChat: "direct_chat",
  GroupChat: "group_chat",
  RoomChat: "room_chat",
  Dialogic: "dialogic",
} as const;

export const InteractionAudience = {
  Private: "private",
  Public: "public",
  Whisper: "whisper",
} as const;

export const InteractionEventType = {
  Speech: "speech",
  Gift: "gift",
  ShowItem: "show_item",
  Trade: "trade",
  Threaten: "threaten",
  Act: "act",
} as const;

export const PlayerIntentActionType = {
  Move: "move",
  Talk: "talk",
  Act: "act",
  UseItem: "use_item",
  Gift: "gift",
  ShowItem: "show_item",
  Trade: "trade",
  Threaten: "threaten",
  Inspect: "inspect",
  Wait: "wait",
  Unknown: "unknown",
} as const;
