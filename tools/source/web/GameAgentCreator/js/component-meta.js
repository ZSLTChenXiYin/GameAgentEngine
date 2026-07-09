window.GAMEAGENT_COMPONENT_META = [
  {
    "type": "profile",
    "display_name": "档案",
    "validation_mode": "weak",
    "data_format": "json_object",
    "help_text": "JSON object required; fields are flexible."
  },
  {
    "type": "rule",
    "display_name": "规则",
    "validation_mode": "free",
    "data_format": "text",
    "help_text": "Free text allowed."
  },
  {
    "type": "timeline",
    "display_name": "时间线",
    "validation_mode": "free",
    "data_format": "text",
    "help_text": "Free text allowed."
  },
  {
    "type": "action_policy",
    "display_name": "动作策略",
    "validation_mode": "free",
    "data_format": "text",
    "help_text": "Free text allowed."
  },
  {
    "type": "relations",
    "display_name": "关系说明",
    "validation_mode": "free",
    "data_format": "text",
    "help_text": "Free text allowed."
  },
  {
    "type": "prompt_profile",
    "display_name": "提示档案",
    "validation_mode": "free",
    "data_format": "text",
    "help_text": "Free text allowed."
  },
  {
    "type": "lore",
    "display_name": "设定",
    "validation_mode": "free",
    "data_format": "text",
    "help_text": "Free text allowed."
  },
  {
    "type": "autonomous",
    "display_name": "自主行为",
    "validation_mode": "strong",
    "data_format": "json_object",
    "help_text": "Structured autonomous config JSON.",
    "required_fields": [
      "enabled",
      "trigger"
    ],
    "enum_fields": {
      "trigger": [
        "manual",
        "world_tick_sync",
        "scheduled"
      ]
    },
    "positive_if_equals": {
      "interval_seconds": {
        "trigger": "scheduled"
      }
    }
  },
  {
    "type": "world_state",
    "display_name": "世界状态",
    "validation_mode": "strong",
    "data_format": "json_object",
    "help_text": "Structured current world state for tick continuity. Optional fields must keep their expected string / string-array / object shapes."
  },
  {
    "type": "story_state",
    "display_name": "剧情状态",
    "validation_mode": "strong",
    "data_format": "json_object",
    "help_text": "Structured current narrative state and unresolved threads. Optional fields must keep their expected string / string-array / object shapes."
  },
  {
    "type": "story_history",
    "display_name": "剧情历史",
    "validation_mode": "strong",
    "data_format": "json_object",
    "help_text": "Structured rolling history of recent story beats. entries must be an array of structured history objects."
  },
  {
    "type": "tick_policy",
    "display_name": "Tick 策略",
    "validation_mode": "strong",
    "data_format": "json_object",
    "help_text": "Structured tick policy and continuity constraints. Optional fields must keep their expected string-array / object shapes."
  },
  {
    "type": "world_time_state",
    "display_name": "世界时间状态",
    "validation_mode": "strong",
    "data_format": "json_object",
    "help_text": "Structured current world time state for engine-managed tick progression."
  },
  {
    "type": "state_snapshot",
    "display_name": "状态快照",
    "validation_mode": "weak",
    "data_format": "json_object",
    "help_text": "Structured snapshot payload for state rollups and checkpoints."
  }
];
