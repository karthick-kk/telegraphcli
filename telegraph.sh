#!/bin/bash
# Simple wrapper script for common telegraphcli operations

# Function to display usage information
usage() {
  echo "Usage: $0 COMMAND [options] [--verbose]"
  echo ""
  echo "Commands:"
  echo "  create-user                Create a new Telegraph user"
  echo "  view-user                  View current user information"
  echo "  edit-user                  Edit user details"
  echo "  revoke-token               Revoke and regenerate access token"
  echo "  create-post FILE TITLE     Create a new post from Markdown file"
  echo "  list-posts                 List all posts for current user"
  echo "  view-post PATH             Get details about a post by its path"
  echo "  edit-post PATH FILE        Edit a post using a Markdown file"
  echo "  delete-post PATH           Delete a post by its path"
  echo "  get-views PATH             Get view count for a post"
  echo "  help                       Show this help"
  echo ""
  echo "Options:"
  echo "  --verbose                  Show detailed debug information"
  exit 1
}

# Ensure command argument is provided
if [ $# -lt 1 ]; then
  usage
fi

# Check for --verbose flag
verbose_arg=""
args=()
for arg in "$@"; do
  if [[ "$arg" == "--verbose" ]]; then
    verbose_arg="--verbose"
  else
    args+=("$arg")
  fi
done

# If --verbose was found, remove it from args for further processing
if [[ -n "$verbose_arg" ]]; then
  # Re-assign positional parameters without --verbose
  set -- "${args[@]}"
fi

# Execute command based on first argument
case "$1" in
  create-user)
    ./telegraphcli user create $verbose_arg
    ;;
  view-user)
    ./telegraphcli user view $verbose_arg
    ;;
  edit-user)
    ./telegraphcli user edit $verbose_arg
    ;;
  revoke-token)
    ./telegraphcli user revoke $verbose_arg
    ;;
  create-post)
    if [ $# -lt 3 ]; then
      echo "Error: create-post requires FILE and TITLE arguments"
      usage
    fi
    ./telegraphcli page create "$2" "$3" $verbose_arg
    ;;
  list-posts)
    ./telegraphcli page list $verbose_arg
    ;;
  view-post)
    if [ $# -lt 2 ]; then
      echo "Error: view-post requires PATH argument"
      usage
    fi
    ./telegraphcli page get "$2" $verbose_arg
    ;;
  edit-post)
    if [ $# -lt 3 ]; then
      echo "Error: edit-post requires PATH and FILE arguments"
      usage
    fi
    ./telegraphcli page edit "$2" "$3" $verbose_arg
    ;;
  delete-post)
    if [ $# -lt 2 ]; then
      echo "Error: delete-post requires PATH argument"
      usage
    fi
    ./telegraphcli page delete "$2" $verbose_arg
    ;;
  get-views)
    if [ $# -lt 2 ]; then
      echo "Error: get-views requires PATH argument"
      usage
    fi
    ./telegraphcli page views "$2" $verbose_arg
    ;;
  help|--help|-h)
    usage
    ;;
  *)
    echo "Unknown command: $1"
    usage
    ;;
esac
