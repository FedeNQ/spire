 #!/usr/bin/env bash
          kind_release_info=$(curl -s https://api.github.com/repos/kubernetes-sigs/kind/releases/latest)
          kind_version=$(echo "$kind_release_info" | jq -r '.tag_name')
          tags=$(curl -s "https://hub.docker.com/v2/repositories/kindest/node/tags" | jq '.results[].name')
          
          readarray -t tags_sorted < <(printf '%s\n' "$tags" | sort -V)

          declare -A tags_map
          for element in "${tags_sorted[@]}"; do
            # Element is in this form: "X.XX.YY"
            # Extract the "X.XX" part as the key for the map
            key="${element%.*}"
            key="${key//\"}"

            # Extract the "YY" part as the value for the map
            value="${element##*.}"
            value="${value//\"}"

            # Because tags_sorted is sorted, we just need to keep the last value seen per key
            tags_map[$key]=$value
           done

          # Read the content of the array.txt file
          IFS= readarray -t matrix_lines < ./.github/k8s-version/array.txt

          # Convert each line of the file into a JSON array element
          json_array="["
          for line in "${matrix_lines[@]}"; do
          json_array+="$line,"
          done

          # Add every version from tags_map
          for key in "${!tags_map[@]}"; do
              value="${tags_map[$key]}"
              k8s_image="kindest/node:$key.$value"
              new_version_row="[\"$key.$value\",\"$k8s_image\",\"$kind_version\"]"
              json_array+="$new_version_row,"
          done

          json_array="${json_array%,}]"

echo ${json_array}