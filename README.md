# Usage
Start the server with one or more geojson files:
    
    ./gopointserver ./stemlokalen.geojson

Fetch all features in a certain bounding box:

    http://localhost:8000/features?bbox=6.546928882598878,53.21316473076784,6.572463512420655,53.220732537799954

Fetch all features in a certain radius of a point:

    http://localhost:8000/nearest?point=6.546928882598878,53.21316473076784&radius=0.01

# Building
I haven't figured out yet how go install works. So for now, just clone the repository and run

    git clone git@github.com/jelmervdl/gopointserver.git
    go build

# Limitations
The server currently only supports points. No other (more complex) geometry is supported, and will cause
the server to crash :)

