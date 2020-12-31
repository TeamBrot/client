import sys
import json
import tempfile
import os
from PIL import Image, ImageDraw
import ffmpeg

SCALING = 16
COLORS = ["#dddddd", "#ff0000", "#00ff00", "#0000ff", "#00ffff", "#ffff00", "#ff00ff", "#000000"]

if len(sys.argv) < 2:
    print("no file given")
    exit(1)

basename = os.path.basename(sys.argv[1]).removesuffix(".json")

with open(sys.argv[1]) as f:
    data = json.load(f)

width = data["game"][0]["width"]
height = data["game"][0]["height"]

with tempfile.TemporaryDirectory() as tmpdir:
    for index, status in enumerate(data["game"]):
        im = Image.new("RGB", (width * SCALING, height * SCALING))
        draw = ImageDraw.Draw(im)
        filename = os.path.join(tmpdir, basename + "-" + str(index).zfill(4) + ".png")
        for i,y in enumerate(status["cells"]):
            for j,x in enumerate(y):
                draw.rectangle([j*SCALING,i*SCALING,(j+1)*SCALING-1,(i+1)*SCALING-1], fill=COLORS[x])
        im.save(filename)
        print(filename)

    (
        ffmpeg
        .input(os.path.join(tmpdir, "*.png"), pattern_type='glob', framerate=10)
        .output(basename + ".mp4")
        .run()
    )
