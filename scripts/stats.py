#!/usr/bin/python3
import os
import sys
import json
from visualize import place
from collections import defaultdict
import argparse

"""
result: {
    client: 'combi',
    players: 2,
    width: 15,
    height: 15,
    deadline: 2,
    offset: 4,

    myStartProbability: 1.2
    minimaxActivationValue: 0.01
    filterValue: 1

    place: 1
}
"""

ATTRIBUTES = [
    "myStartProbability",
    "minimaxActivationValue",
    "filterValue"
]

CLIENTS =[
    "combi",
    "minimax",
    "rollouts",
    "probability",
    "smart"
]

def has_errors(game_path):
    for error in os.scandir(os.path.join(game_path, "error")):
        if not error.is_file():
            continue
        if os.path.getsize(error.path) != 0:
            return True
    return False

def file_result(file_path: str):
    print("reading", file_path)
    with open(file_path) as f:
        data = json.load(f)

    return {
        "client": data["config"]["clientName"],
        "players": len(data["game"][0]["players"]),
        "width": data["game"][0]["width"],
        "height": data["game"][0]["height"],
        "deadline": 0,
        "offset": 0,
        "myStartProbability": data["config"]["myStartProbability"],
        "minimaxActivationValue": data["config"]["minimaxActivationValue"],
        "filterValue": data["config"]["filterValue"],
        "place": place(data)
    }

def dir_result(game_path: str):
    results = []
    for result_file in os.scandir(game_path):
        if not result_file.is_file or not result_file.path.endswith(".json"):
            continue
        results.append(file_result(result_file.path))

    return results

def get_results_from_dir(directory: str):
    results = []
    total_games = 0
    error_games = 0
    for game in os.scandir(directory):
        total_games += 1
        if has_errors(game.path):
            print("skipping", game.path, "since it contains errors")
            error_games += 1
            continue
        try:
            results += dir_result(game.path)
        except Exception as e:
            print("could not get game results:", e)
            print("skipping", game.path)

    print("out of", total_games, "games,", error_games, "contain errors")
    return results

def parse_args():
    parser = argparse.ArgumentParser(description="process logfiles and show statistics")
    parser.add_argument("files", metavar="file", help="json logfiles or directory structures", nargs='*')
    parser.add_argument("--output", "-o", help="diagram output file, can be jpeg, png, svg")
    return parser.parse_args()

if __name__ == '__main__':
    args = parse_args()
    results = []
    for f in args.files:
        if f.endswith(".json"):
            results.append(file_result(f))
        else:
            results += get_results_from_dir(f)

    import matplotlib.pyplot as plt
    fig, ax = plt.subplots(1, 3, figsize=(10,5))
    fig.tight_layout()

    for i, attribute in enumerate(ATTRIBUTES):
        wins = defaultdict(int)
        total_games = defaultdict(int)
        win_percentage = dict()
        print("results for", attribute)
        for result in results:
            total_games[result[attribute]] += 1
            if result["place"] == 1:
                wins[result[attribute]] += 1

        for value in sorted(map(float, total_games.keys())):
            win_percentage[value] = wins[value] / total_games[value]

            print(value, "\t", total_games[value], wins[value], "{:.3f}".format(win_percentage[value]))

        x = win_percentage.keys()
        y = [ win_percentage[key] for key in x ]
        x = list(map(str, x))
        ax[i].bar(x, y)
        ax[i].set_xlabel(attribute)

    if args.output is not None:
        plt.savefig(args.output, format=args.output.split('.')[-1])

    for i, client in enumerate(CLIENTS):
        wins = 0
        total_games = 0
        for result in results:
            if result["client"] == client:
                total_games += 1
                if result["place"] == 1:
                    wins += 1

        if total_games == 0:
            continue
        win_percentage = wins / total_games
        print("results for", client)
        print(total_games, wins, "{:.3f}".format(win_percentage))
