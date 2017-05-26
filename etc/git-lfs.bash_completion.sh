# git-lfs tab-completion script for bash.
# This script complements the completion script that ships with git.

# Check that git tab completion is available
if declare -F _git > /dev/null; then
  # Duplicate and rename the 'list_all_commands' function
  eval "$(declare -f __git_list_all_commands | \
        sed 's/__git_list_all_commands/__git_list_all_commands_without_git-lfs/')"

  # Wrap the 'list_all_commands' function with extra hub commands
  __git_list_all_commands() {
    cat <<-EOF
lfs
EOF
    __git_list_all_commands_without_git-lfs
  }

  # Ensure cached commands are cleared
  __git_all_commands=""

  ##########################
  # git-lfs command completions
  ##########################

  _git_lfs() {
    local cmds="clean env init logs ls-files push smudge status track untrack"
    if [[ ${COMP_CWORD} -eq 2 ]] ; then
      __gitcomp "$cmds"
      return 0
    elif [[ ${COMP_CWORD} -eq 3 ]] ; then
      case "${prev}" in
        clean|smudge)
          _filedir -f
          return
          ;;
        logs)
          sub_cmds='--clear --boomtown'
          _git_branch
          ;;
        ls-files)
          sub_cmds="$(__git_refs)"
          ;;
        push)
          sub_cmds="--dry-run --stdin $(__git_refs)"
          ;;
        status)
          sub_cmds="--porcelain"
          ;;
        *)
          sub_cmds=''
          ;;
      esac
      __gitcomp "$sub_cmds"
      return
    elif [[ ${COMP_CWORD} -eq 4 ]] ; then
      case "${prev}" in
        --boomtown|--clear)
          _git_branch
        ;;
      esac
    fi
  }
fi
