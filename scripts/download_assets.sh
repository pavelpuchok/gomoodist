set -xe

ARCHIVE_URL="https://github.com/remvze/moodist/archive/4c8d5775274ad0573c73e30e5aae4fc87361e0e9.zip"
ARCHIVE_NAME="moodist.zip"

DEST_DIR=$(mktemp -d)

# Download archive
curl -L "$ARCHIVE_URL" -o "$DEST_DIR/$ARCHIVE_NAME"

# Extract
unzip "$DEST_DIR/$ARCHIVE_NAME" -d "$DEST_DIR"

# Get top-level extracted dir name (GitHub zip adds a subdir)
EXTRACTED_DIR=$(find "$DEST_DIR" -mindepth 1 -maxdepth 1 -type d)

# Convert mp3 & wav to ogg
mapfile -d '' files < <(find "$EXTRACTED_DIR/public/sounds/" -type f \( -name "*.mp3" -o -name "*.wav" \) -print0)
for file in "${files[@]}"; do
  ffmpeg -loglevel fatal -y -i "$file" -c:a libvorbis -ar 44100 -ac 2 -qscale:a 6 "${file%.*}.ogg"
  rm "$file"
done

# Clean old files
rm -r internal/app/assets/sounds || true

# Move files from temp dir to assets
mv "$EXTRACTED_DIR/public/sounds" ./internal/app/assets/

# Clean up temp dir
rm -rf "$DEST_DIR"
