#!/usr/bin/env bash

set -e

MODEL="huihui_ai/Qwen3.8-abliterated:27b"
HOST="http://localhost:11434"

# Start the prewarm process in the background.
(
    sleep 2

    until curl -sf "${HOST}/api/tags" >/dev/null; do
        sleep 1;
    done

    echo "Prewarming ${MODEL} ...";

    curl -sf "${HOST}/api/generate" -d "{\"model\":\"${MODEL}\",\"keep_alive\":-1}" >/dev/null;

    echo "${MODEL} prewarmed.";

) &

# Keep Ollama in the foreground.
# export OLLAMA_CONTEXT_LENGTH=32768;
# export OLLAMA_CONTEXT_LENGTH=131072;

export OLLAMA_CONTEXT_LENGTH=262144;
export OLLAMA_KEEP_ALIVE=-1;
exec ollama serve;

