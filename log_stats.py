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

ATTRIBUTES = [
    "myStartProbability",
    "minimaxActivationValue",
    "filterValue"
]

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
        try:
            results += get_results(game.path)
        except Exception as e:
            print("could not get game results:", e)
            print("skipping", game.path)


    print("out of", total_games, "games,", error_games, "contain errors")

    import matplotlib.pyplot as plt
    fig, ax = plt.subplots(1, 3, figsize=(10,5))
    # fig.suptitle("Gewinnhäufigkeit in Abhängigkeit verschiedener Hyperparameter")
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

        for value in sorted(total_games.keys()):
            win_percentage[value] = wins[value] / total_games[value]

            print(value, "\t", total_games[value], wins[value], "{:.3f}".format(win_percentage[value]))

        x = win_percentage.keys()
        y = [ win_percentage[key] for key in x ]
        x = list(map(str, x))
        ax[i].bar(x, y)
        ax[i].set_xlabel(attribute)
    plt.savefig('hyperparameters.svg', format="svg")
