# Usage
Start the server with one or more geojson files:
  ./gopointserver 8080 ./stemlokalen.geojson

Now it will answer requests such as:
  http://localhost:8000/?bbox=6.546928882598878,53.21316473076784,6.572463512420655,53.220732537799954

# Building
I haven't figured out yet how go install works. So for now, just clone the repository and run
  go build

# Limitations
The server currently only supports points. No other (more complex) geometry is supported, and will cause
the server to crash :)

