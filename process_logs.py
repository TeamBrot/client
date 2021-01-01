import sys
import json
import tempfile
import os
from PIL import Image, ImageDraw
import ffmpeg

SCALING = 16
COLORS = ["#dddddd", "#ff0000", "#00ff00", "#0000ff", "#00ffff", "#ffff00", "#ff00ff", "#000000"]
NUM_FINAL_FRAMES = 30
FPS = 10

def make_image(status, filename):
    width = status["width"]
    height = status["height"]
    im = Image.new("RGB", (width * SCALING, height * SCALING))
    draw = ImageDraw.Draw(im)
    for i,y in enumerate(status["cells"]):
        for j,x in enumerate(y):
            draw.rectangle([j*SCALING,i*SCALING,(j+1)*SCALING-1,(i+1)*SCALING-1], fill=COLORS[x])
    im.save(filename)
    print(filename)


if len(sys.argv) < 2:
    print("no file given")
    exit(1)

basename = os.path.basename(sys.argv[1]).removesuffix(".json")

with open(sys.argv[1]) as f:
    data = json.load(f)


with tempfile.TemporaryDirectory() as tmpdir:
    index = 0
    for status in data["game"]:
        filename = os.path.join(tmpdir, basename + "-" + str(index).zfill(4) + ".png")
        make_image(status, filename)
        index += 1
    for _ in range(NUM_FINAL_FRAMES):
        filename = os.path.join(tmpdir, basename + "-" + str(index).zfill(4) + ".png")
        make_image(data["game"][-1], filename)
        index += 1

    (
        ffmpeg
        .input(os.path.join(tmpdir, "*.png"), pattern_type='glob', framerate=FPS)
        .output(basename + ".mp4")
        .run()
    )
