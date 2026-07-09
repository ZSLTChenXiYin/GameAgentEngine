window.GAMEAGENT_COMPONENT_META = [
  {
    "type": "profile",
    "validation_mode": "weak",
    "data_format": "json_object",
    "help_text": "JSON object required; fields are flexible."
  },
  {
    "type": "rule",
    "validation_mode": "free",
    "data_format": "text",
    "help_text": "Free text allowed."
  },
  {
    "type": "timeline",
    "validation_mode": "free",
    "data_format": "text",
    "help_text": "Free text allowed."
  },
  {
    "type": "action_policy",
    "validation_mode": "free",
    "data_format": "text",
    "help_text": "Free text allowed."
  },
  {
    "type": "relations",
    "validation_mode": "free",
    "data_format": "text",
    "help_text": "Free text allowed."
  },
  {
    "type": "prompt_profile",
    "validation_mode": "free",
    "data_format": "text",
    "help_text": "Free text allowed."
  },
  {
    "type": "lore",
    "validation_mode": "free",
    "data_format": "text",
    "help_text": "Free text allowed."
  },
  {
    "type": "autonomous",
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
    "validation_mode": "strong",
    "data_format": "json_object",
    "help_text": "Structured current world state for tick continuity. Optional fields must keep their expected string / string-array / object shapes."
  },
  {
    "type": "story_state",
    "validation_mode": "strong",
    "data_format": "json_object",
    "help_text": "Structured current narrative state and unresolved threads. Optional fields must keep their expected string / string-array / object shapes."
  },
  {
    "type": "story_history",
    "validation_mode": "strong",
    "data_format": "json_object",
    "help_text": "Structured rolling history of recent story beats. entries must be an array of structured history objects."
  },
  {
    "type": "tick_policy",
    "validation_mode": "strong",
    "data_format": "json_object",
    "help_text": "Structured tick policy and continuity constraints. Optional fields must keep their expected string-array / object shapes."
  },
  {
    "type": "world_time_state",
    "validation_mode": "strong",
    "data_format": "json_object",
    "help_text": "Structured current world time state for engine-managed tick progression."
  },
  {
    "type": "state_snapshot",
    "validation_mode": "weak",
    "data_format": "json_object",
    "help_text": "Structured snapshot payload for state rollups and checkpoints."
  }
];
