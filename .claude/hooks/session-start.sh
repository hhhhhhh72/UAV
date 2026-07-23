#!/usr/bin/env bash
# SessionStart hook: injects superpowers bootstrap skill
set -euo pipefail

SKILL_FILE="d:/w-yao/.claude/skills/using-superpowers/SKILL.md"
if [ ! -f "$SKILL_FILE" ]; then
  SKILL_FILE="d:/w-yao/.claude/plugins/superpowers/skills/using-superpowers/SKILL.md"
fi

CONTENT=$(cat "$SKILL_FILE" 2>/dev/null || echo "Error reading skill file")

# JSON-escape the content
escape_for_json() {
  local s="$1"
  s="${s//\\/\\\\}"
  s="${s//\"/\\\"}"
  s="${s//$'\n'/\\n}"
  s="${s//$'\r'/\\r}"
  s="${s//$'\t'/\\t}"
  printf '%s' "$s"
}

ESCAPED=$(escape_for_json "$CONTENT")

# Build JSON in parts to avoid printf interpreting escape sequences in content
PREFIX='{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"'
HEADER='<EXTREMELY_IMPORTANT>\nYou have superpowers.\n\n**Below is the full content of your '"'using-superpowers'"' skill - your introduction to using skills. For all other skills, use the '"'Skill'"' tool:**\n\n'
SUFFIX='\n</EXTREMELY_IMPORTANT>"}}'

printf '%s%s%s%s' "$PREFIX" "$HEADER" "$ESCAPED" "$SUFFIX"
echo
