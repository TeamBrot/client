#!/usr/bin/python3
import sys
import json
import tempfile
import os
from PIL import Image, ImageDraw, ImageFont
import ffmpeg
from common import place

SCALING = 64
COLORS = ["#dddddd", "#ff0000", "#00ff00", "#0000ff",
          "#00ffff", "#ffff00", "#ff00ff", "#000000"]
PLAYER_COLOR_INDEX = 1
FPS = 30
NUM_FINAL_FRAMES = 30
NUM_START_FRAMES = 30


def json_basename(json_filename):
    return os.path.basename(json_filename).removesuffix(".json")


def video_filename(json_filename):
    return json_filename.removesuffix(".json") + ".mp4"


def image_filename(tmpdir, json_filename, index):
    return os.path.join(tmpdir, json_basename(json_filename) + "-" + str(index).zfill(4) + ".png")


def start_image(data, colors, font):
    width = data["game"][0]["width"]
    height = data["game"][0]["height"]
    client_name = data["config"]["clientName"]
    game_url = data["config"]["gameURL"]
    start_time = data["start"][:19]
    numplayers_start = len(data["game"][0]["players"])
    text = (
        "time: " + start_time + "\n"
        "server: " + game_url + "\n"
        "client: " + client_name + "\n\n"
        "width: " + str(width) + "\n"
        "height: " + str(height) + "\n"
        "number of players: " + str(numplayers_start) + "\n\n"
        "place: " + str(place(data)) + "\n\n"
        "our color:"
    )
    im = Image.new("RGB", (width * SCALING, height * SCALING + 100))
    draw = ImageDraw.Draw(im)
    draw.text((10, 10), text, font=font)
    draw.rectangle([150, 330, 150+2*16, 330+2*16],
                   colors[data["game"][0]["you"]])
    return im

def draw_square(draw, i, j, color):
    draw.rectangle([j*SCALING, i*SCALING+40, (j+1)*SCALING-1, (i+1)*SCALING-1+40], fill=color)

def create_image(width, height):
    return Image.new("RGB", (width * SCALING, height * SCALING + 40))

def board_image(status, colors, font, turn):
    width = status["width"]
    height = status["height"]
    im = create_image(width, height)
    draw = ImageDraw.Draw(im)
    draw.text((0, 0), "Turn {}".format(turn), font=font)
    for i, y in enumerate(status["cells"]):
        for j, x in enumerate(y):
            draw.rectangle([j*size+x_offset, i*size+y_offset, (j+1) *
                            size + x_offset - 1, (i + 1) * size + y_offset - 1], fill=colors[x], outline="black", width=1)
    for n in status["players"]:
        player = status["players"][n]
        if player["active"]:
            draw.rectangle([player["x"]*size +
                            x_offset, player["y"]*size+y_offset, (player["x"]+1)*size+x_offset-1, (player["y"]+1)*size+y_offset-1], fill=HEADCOLORS[int(n)-1])
    return im


def make_video(json_filename):
    # basename = json_basename(json_filename)
    font = ImageFont.truetype("arial.ttf", size=30)
    with open(json_filename) as f:
        data = json.load(f)
    player_id = data["game"][0]["you"]
    colors = [color for color in COLORS]
    colors[PLAYER_COLOR_INDEX], colors[player_id] = colors[player_id], colors[PLAYER_COLOR_INDEX]
    index = 0
    with tempfile.TemporaryDirectory() as tmpdir:
        index = 0
        im = start_image(data, colors, font)
        for _ in range(NUM_START_FRAMES):
            im.save(image_filename(tmpdir, json_filename, index))
            index += 1
        for turn, status in enumerate(data["game"]):
            im = board_image(status, colors, font, turn+1)
            im.save(image_filename(tmpdir, json_filename, index))
            index += 1
        for _ in range(NUM_FINAL_FRAMES):
            im.save(image_filename(tmpdir, json_filename, index))
            index += 1

        (
            ffmpeg
            .input(os.path.join(tmpdir, "*.png"), pattern_type='glob', framerate=FPS)
            .output(video_filename(json_filename), vcodec='libx264')
            .global_args('-loglevel', 'error')
            .run()
        )


if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("no filenames given")
        exit(1)

    for json_filename in sys.argv[1:]:
        output_filename = video_filename(json_filename)
        if not os.path.exists(output_filename):
            print("processing " + json_filename + "...")
            try:
                make_video(json_filename)
            except:
                print("couldn't process this file. probably the game ended unnormaly")
            print("wrote to", output_filename)
        else:
            print("skipping", json_filename, "because output",
                  output_filename, "already exists...")
