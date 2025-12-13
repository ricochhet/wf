cd submodules/warframe-public-export
find . -type f -name "*.json" -exec bash -c '
  for f; do
    rel="${f#./}"
    mkdir -p "../../assets/$(dirname "$rel")"
    cp "$f" "../../assets/$rel"
  done
' bash {} +