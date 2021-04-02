#!/usr/bin/python3
import sys
import json
import tempfile
import os
import argparse
from PIL import Image, ImageDraw, ImageFont
import ffmpeg
from common import place

COLORS = ["#ffffff", "#f6a800",  "#4c5b5c", "#98ce00",
          "#a44a3f", "#ff751b", "#ff00ff", "#ededed"]
HEADCOLORS = ["#ffffff", "#bb8000", "#313b3c", "#6d9300",
              "#79372f", "#df5800", "#c400c4", "ededed"]
PLAYER_COLOR_INDEX = 1

# Video dimensions
WIDTH = 1920
HEIGHT = 1080
# Video fps
FPS = 30
# Frames per board position
FPB = 1
# Start and end lengths
START_SEC = 2
END_SEC = 1
# Outline width
OUTLINE = 1


def json_basename(json_filename):
    return os.path.basename(json_filename).removesuffix(".json")


def video_filename(json_filename):
    return json_filename.removesuffix(".json") + ".mp4"


def image_filename(tmpdir, json_filename, index):
    return os.path.join(tmpdir, json_basename(json_filename) + "-" + str(index).zfill(4) + ".png")


def start_image(data, colors, width=WIDTH, height=HEIGHT):
    font = ImageFont.truetype("arial.ttf", size=30)
    board_width = data["game"][0]["width"]
    board_height = data["game"][0]["height"]
    client_name = data["config"]["clientName"]
    game_url = data["config"]["gameURL"]
    start_time = data["start"][:19]
    numplayers_start = len(data["game"][0]["players"])
    text = (
        "time: " + start_time + "\n"
        "server: " + game_url + "\n"
        "client: " + client_name + "\n\n"
        "width: " + str(board_width) + "\n"
        "height: " + str(board_height) + "\n"
        "number of players: " + str(numplayers_start) + "\n\n"
        "place: " + str(place(data)) + "\n\n"
        "our color:"
    )
    im = Image.new("RGB", (width, height), "white")
    draw = ImageDraw.Draw(im)
    draw.text((width/2, height/2), text, anchor="mm",
              align="center", font=font, fill="black")
    draw.rectangle([1030, 685, 1062, 717],
                   colors[data["game"][0]["you"]])
    return im

def draw_square(draw, i, j, color):
    draw.rectangle([j*SCALING, i*SCALING+40, (j+1)*SCALING-1, (i+1)*SCALING-1+40], fill=color)

def create_image(width, height):
    return Image.new("RGB", (width * SCALING, height * SCALING + 40))

def board_image(status, colors, headcolors, width=WIDTH, height=HEIGHT, outline=OUTLINE):
    board_width = status["width"]
    board_height = status["height"]
    size = math.floor(height / board_height)

    if size * board_width > width:
        size = math.floor(width / board_width)

    x_offset = (width-size*board_width)/2
    y_offset = (height-size*board_height)/2
    im = Image.new("RGB", (width, height), "white")
    draw = ImageDraw.Draw(im)
    draw.text((0, 0), "Turn {}".format(turn), font=font)
    for i, y in enumerate(status["cells"]):
        for j, x in enumerate(y):
            draw.rectangle([j*size+x_offset, i*size+y_offset, (j+1)*size+x_offset-1, (i+1)*size+y_offset-1], fill=colors[x], outline="black", width=outline)
    for n in status["players"]:
        player = status["players"][n]
        if player["active"]:
            draw.rectangle([player["x"]*size+x_offset, player["y"]*size+y_offset, (player["x"]+1)*size+x_offset-1, (player["y"]+1)*size+y_offset-1], fill=headcolors[int(n)], outline='black', width=outline)
    return im


def make_video(json_filename, width=WIDTH, height=HEIGHT, fps=FPS, fpb=FPB, start_frames=FPS*START_SEC, end_frames=FPS*END_SEC, outline=OUTLINE):
    with open(json_filename) as f:
        data = json.load(f)
    player_id = data["game"][0]["you"]

    # Adjust colors to current player so that current player is always colors[1]
    colors = [color for color in COLORS]
    colors[PLAYER_COLOR_INDEX], colors[player_id] = colors[player_id], colors[PLAYER_COLOR_INDEX]
    headcolors = [headcolor for headcolor in HEADCOLORS]
    headcolors[PLAYER_COLOR_INDEX], headcolors[player_id] = headcolors[player_id], headcolors[PLAYER_COLOR_INDEX]

    index = 0
    with tempfile.TemporaryDirectory() as tmpdir:
        index = 0
        im = start_image(data, colors, width=width, height=height)
        for _ in range(start_frames):
            im.save(image_filename(tmpdir, json_filename, index))
            index += 1
        for status in data["game"]:
            im = board_image(status, colors, headcolors, width=width, height=height, outline=outline)
            for _ in range(fpb):
                im.save(image_filename(tmpdir, json_filename, index))
                index += 1
        for _ in range(end_frames):
            im.save(image_filename(tmpdir, json_filename, index))
            index += 1

        (
            ffmpeg
            .input(os.path.join(tmpdir, "*.png"), pattern_type='glob', framerate=fps)
            .output(video_filename(json_filename))
            .global_args('-loglevel', 'error', '-y')
            .run()
        )


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Visualize spe_ed JSON game logs as videos')
    parser.add_argument('files', nargs='+', help='log files to visualize')
    parser.add_argument('--fps', type=int, default=FPS, help='frames per second')
    parser.add_argument('--fpb', type=int, default=FPB, help='frames per board position')
    parser.add_argument('--start', type=float, default=START_SEC, help='number of seconds the start image is shown')
    parser.add_argument('--end', type=float, default=END_SEC, help='number of seconds the end image is shown')
    parser.add_argument('--width', type=int, default=WIDTH, help='video width in pixels')
    parser.add_argument('--height', type=int, default=HEIGHT, help='video height in pixels')
    parser.add_argument('--outline', type=int, default=OUTLINE, help='outline width in pixels')
    parser.add_argument('--force', '-f', action='store_true', help='overwrite existing video file')
    args = parser.parse_args()

    for json_filename in args.files:
        output_filename = video_filename(json_filename)
        if args.force or not os.path.exists(output_filename):
            print("processing " + json_filename + "...")
            make_video(json_filename, width=args.width, height=args.height, fps=args.fps, fpb=args.fpb, start_frames=args.fps*args.start, end_frames=args.fps*args.end, outline=args.outline)
            print("wrote to", output_filename)
        else:
            print("skipping", json_filename, "because output",
                  output_filename, "already exists...")
