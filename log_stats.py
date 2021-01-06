#!/usr/bin/python3
import os
import sys
import json
from process_logs import place
from collections import defaultdict

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

ATTRIBUTE = "filterValue"

def has_errors(game_path):
    for error in os.scandir(os.path.join(game_path, "error")):
        if not error.is_file():
            continue
        if os.path.getsize(error.path) != 0:
            return True
    return False

def get_results(game_path: str):
    results = []
    for result_file in os.scandir(game_path):
        if not result_file.is_file or not result_file.path.endswith(".json"):
            continue

        with open(result_file.path) as f:
            data = json.load(f)

        results.append({
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
        })
    return results



if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("dir required")
        exit()
    directory = sys.argv[1]

    results = []
    total_games = 0
    error_games = 0
    for game in os.scandir(directory):
        total_games += 1
        if has_errors(game.path):
            print("skipping", game.path, "since it contains errors")
            error_games += 1
            continue
        print("reading", game.path)
        results += get_results(game.path)

    print("out of", total_games, "games,", error_games, "contain errors")

    wins = defaultdict(int)
    total_games = defaultdict(int)
    win_percentage = {}

    print("results for", ATTRIBUTE)
    for result in results:
        total_games[result[ATTRIBUTE]] += 1
        if result["place"] == 1:
            wins[result[ATTRIBUTE]] += 1

    for value in sorted(total_games.keys()):
        win_percentage[value] = wins[value] / total_games[value]

        print(value, "\t", total_games[value], wins[value], win_percentage[value])