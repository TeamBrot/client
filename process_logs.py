#!/usr/bin/python3
import sys
import json
import tempfile
import os
from PIL import Image, ImageDraw, ImageFont
import ffmpeg

SCALING = 16
COLORS = ["#dddddd", "#ff0000", "#00ff00", "#0000ff", "#00ffff", "#ffff00", "#ff00ff", "#000000"]
PLAYER_COLOR_INDEX = 1
FPS = 10
NUM_FINAL_FRAMES = 30
NUM_START_FRAMES = 20

def json_basename(json_filename):
    return os.path.basename(json_filename).removesuffix(".json")

def video_filename(json_filename):
    return json_filename.removesuffix(".json") + ".mp4"

def image_filename(tmpdir, json_filename, index):
    return os.path.join(tmpdir, json_basename(json_filename) + "-" + str(index).zfill(4) + ".png")

def start_image(data, colors):
    font = ImageFont.truetype("arial.ttf", size=30)
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
    im = Image.new("RGB", (width * SCALING, height * SCALING))
    draw = ImageDraw.Draw(im)
    draw.text((10,10), text, font=font)
    draw.rectangle([150,330,150+2*SCALING,330+2*SCALING], colors[data["game"][0]["you"]])
    return im
    

def board_image(status, colors):
    width = status["width"]
    height = status["height"]
    im = Image.new("RGB", (width * SCALING, height * SCALING))
    draw = ImageDraw.Draw(im)
    for i,y in enumerate(status["cells"]):
        for j,x in enumerate(y):
            draw.rectangle([j*SCALING,i*SCALING,(j+1)*SCALING-1,(i+1)*SCALING-1], fill=colors[x])
    return im

def place(data):
    last_status = data["game"][-1]
    players_end = list(map(lambda player: int(player[0]), filter(lambda player: player[1]["active"], last_status["players"].items())))
    you = last_status["you"]
    # We are active in the last status, so we won
    if you in players_end:
        assert(not last_status["running"])
        return 1
    # We are not active in the last status
    if len(list(players_end)) == 0:
        # We and our enemies dies simultaneously
        assert(not last_status["running"])
        return 1
    return len(players_end) + 1

def make_video(json_filename):
    # basename = json_basename(json_filename)
    with open(json_filename) as f:
        data = json.load(f)
    player_id = data["game"][0]["you"]
    colors = [ color for color in COLORS ]
    colors[PLAYER_COLOR_INDEX], colors[player_id] = colors[player_id], colors[PLAYER_COLOR_INDEX]
    index = 0
    with tempfile.TemporaryDirectory() as tmpdir:
        index = 0
        im = start_image(data, colors)
        for _ in range(NUM_START_FRAMES):
            im.save(image_filename(tmpdir, json_filename, index))
            index += 1
        for status in data["game"]:
            im = board_image(status, colors)
            im.save(image_filename(tmpdir, json_filename, index))
            index += 1
        for _ in range(NUM_FINAL_FRAMES):
            im.save(image_filename(tmpdir, json_filename, index))
            index += 1

        (
            ffmpeg
            .input(os.path.join(tmpdir, "*.png"), pattern_type='glob', framerate=FPS)
            .output(video_filename(json_filename))
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
            make_video(json_filename)
            print("wrote to", output_filename)
        else:
            print("skipping", json_filename, "because output", output_filename, "already exists...")
