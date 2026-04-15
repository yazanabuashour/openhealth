#!/usr/bin/env bash

set -euo pipefail

fail() {
  printf 'error: %s\n' "$1" >&2
  exit 1
}

trim() {
  local value="$1"
  value="${value#"${value%%[![:space:]]*}"}"
  value="${value%"${value##*[![:space:]]}"}"
  printf '%s' "$value"
}

unquote() {
  local value="$1"
  if [[ "$value" == \"*\" && "$value" == *\" ]]; then
    value="${value:1:${#value}-2}"
  elif [[ "$value" == \'*\' && "$value" == *\' ]]; then
    value="${value:1:${#value}-2}"
  fi
  printf '%s' "$value"
}

extract_scalar() {
  local field="$1"
  local file="$2"

  awk -F ':' -v key="$field" '
    $0 ~ ("^" key ":[[:space:]]*") {
      sub("^" key ":[[:space:]]*", "", $0)
      print
      exit
    }
  ' "$file"
}

if [[ $# -ne 1 ]]; then
  fail "usage: scripts/validate-agent-skill.sh <skill-directory>"
fi

skill_dir="${1%/}"
skill_file="$skill_dir/SKILL.md"

[[ -d "$skill_dir" ]] || fail "skill directory not found: $skill_dir"
[[ -f "$skill_file" ]] || fail "missing required file: $skill_file"

first_line="$(sed -n '1p' "$skill_file")"
[[ "$first_line" == "---" ]] || fail "$skill_file must start with YAML frontmatter delimited by ---"

closing_line="$(awk 'NR > 1 && $0 == "---" { print NR; exit }' "$skill_file")"
[[ -n "$closing_line" ]] || fail "$skill_file must include a closing --- line for YAML frontmatter"
(( closing_line > 2 )) || fail "$skill_file frontmatter must contain at least the required fields"

frontmatter_file="$(mktemp)"
trap 'rm -f "$frontmatter_file"' EXIT
sed -n "2,$((closing_line - 1))p" "$skill_file" > "$frontmatter_file"

name="$(extract_scalar "name" "$frontmatter_file")"
name="$(trim "$(unquote "$name")")"
[[ -n "$name" ]] || fail "$skill_file frontmatter must define a non-empty name"

parent_dir="$(basename "$skill_dir")"
[[ "$name" == "$parent_dir" ]] || fail "$skill_file name must match the parent directory ('$parent_dir')"

(( ${#name} <= 64 )) || fail "$skill_file name must be 64 characters or fewer"
[[ "$name" =~ ^[a-z0-9]+(-[a-z0-9]+)*$ ]] || fail "$skill_file name must use lowercase letters, numbers, and single hyphens only"

description="$(extract_scalar "description" "$frontmatter_file")"
description="$(trim "$(unquote "$description")")"
[[ -n "$description" ]] || fail "$skill_file frontmatter must define a non-empty description"
(( ${#description} <= 1024 )) || fail "$skill_file description must be 1024 characters or fewer"

if grep -q '^compatibility:' "$frontmatter_file"; then
  compatibility="$(extract_scalar "compatibility" "$frontmatter_file")"
  compatibility="$(trim "$(unquote "$compatibility")")"
  [[ -n "$compatibility" ]] || fail "$skill_file compatibility must be non-empty when provided"
  (( ${#compatibility} <= 500 )) || fail "$skill_file compatibility must be 500 characters or fewer"
fi

printf 'validated %s\n' "$skill_dir"
