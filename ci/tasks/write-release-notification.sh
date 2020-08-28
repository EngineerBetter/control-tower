#!/bin/bash

new_release=$(jq --compact-output '.[]' new-versions/release-versions.json)
old_release=$(cat old-versions/release-versions.json)
cup_ref="$(cat control-tower/.git/ref)"
ops_ref="$(cat control-tower-ops/.git/ref)"
cup_message="$(cat control-tower/.git/commit_message)"
ops_message="$(cat control-tower-ops/.git/commit_message)"

touch slack-message/text
cat << EOF > slack-message/text
Control-Tower is ready for a new release, all system tests passed.
EOF

for component in $new_release; do
  name=$(echo "$component" | jq --raw-output '.name')
  new_version=$(echo "$component" | jq --raw-output '.version')
  old_version=$(echo "$old_release" | jq --raw-output --arg name $name '.[] | select(.name==$name).version')
  
  if [ "$(printf '%s\n' "$new_version" "$old_version" | sort -V | head -n1)" != "$new_version" ]; then 
    echo "$name: $old_version > $new_version" >> slack-message/text
  fi
done

cat << EOF >> slack-message/text
Latest commit in *control-tower* repository: \`$cup_ref\`
\`\`\`$cup_message\`\`\`

Latest commit in *control-tower-ops* repository: \`$ops_ref\`
\`\`\`$ops_message\`\`\`
EOF
