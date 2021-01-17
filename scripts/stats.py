#!/usr/bin/python3
import os
import sys
import json
from common import place
from collections import defaultdict
import argparse

"""
result: {
    startTime: "2021-01-01..."
    client: 'combi',
    players: 2,
    width: 15,
    height: 15,
    deadline: 2,
    offset: 4,

    occupiedRatio: 0.1241

    myStartProbability: 1.2
    minimaxActivationValue: 0.01
    filterValue: 1

    place: 1
    numPlayers: 5

    enemyNames: set("name",...)
    ourName: "name"
    endActiveEnemies: set("name")
    
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
    "basic"
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

    you = data["game"][-1]["you"]

    total_fields = data["game"][0]["width"] * data["game"][0]["height"]
    occupied_fields = sum([0 if j == 0 else 1 for i in data["game"][-1]["cells"] for j in i])

    names = set()
    end_active_enemies = set()
    for player in data["game"][-1]["players"]:
        if int(player) == int(you):
            continue
        name = data["game"][-1]["players"][player]["name"]
        if name != '':
            names.add(name)
            if data["game"][-1]["players"][player]["active"]:
                end_active_enemies.add(name)

    if place(data) != 1 and len(end_active_enemies) == 0:
        for player in data["game"][-2]["players"]:
            if int(player) == int(you):
                continue
            name = data["game"][-2]["players"][player]["name"]
            if name != '':
                if data["game"][-2]["players"][player]["active"]:
                    end_active_enemies.add(name)

    return {
        "startTime": data["start"],
        "client": data["config"]["clientName"],
        "players": len(data["game"][0]["players"]),
        "width": data["game"][0]["width"],
        "height": data["game"][0]["height"],
        "deadline": 0,
        "offset": 0,
        "myStartProbability": data["config"]["myStartProbability"],
        "minimaxActivationValue": data["config"]["minimaxActivationValue"],
        "filterValue": data["config"]["filterValue"],
        "place": place(data),
        "occupiedRatio": occupied_fields / total_fields,
        "enemyNames": names,
        "ourName": data["game"][-1]["players"][str(you)]["name"],
        "numPlayers": len(data["game"][0]["players"]),
        "endActiveEnemies": end_active_enemies
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

def get_results(file_or_dir_names):
    results = []
    for f in file_or_dir_names:
        if f.endswith(".json"):
            results.append(file_result(f))
        else:
            results += get_results_from_dir(f)
    return results

def get_client_stats(results):
    total_games = defaultdict(int)
    won_games = defaultdict(int)
    win_percentage = {}
    for result in results:
        client = result["client"]
        total_games[client] += 1
        if result["place"] == 1:
            won_games[client] += 1
            
    for client in total_games:
        win_percentage[client] = won_games[client] / total_games[client]
    return total_games, won_games, win_percentage

def get_attribute_stats(results):
    wins = defaultdict(lambda: defaultdict(int))
    total_games = defaultdict(lambda: defaultdict(int))
    for result in results:
        for attribute in ATTRIBUTES:
            total_games[attribute][result[attribute]] += 1
            if result["place"] == 1:
                wins[attribute][result[attribute]] += 1

    win_percentage = {attribute: {value: wins[attribute][value] / total_games[attribute][value] for value in total_games[attribute].keys()} for attribute in ATTRIBUTES}
    return total_games, wins, win_percentage

def create_attribute_diagram(results, output):
    _, _, win_percentage = get_attribute_stats(results)
    import matplotlib.pyplot as plt
    fig, ax = plt.subplots(1, 3, figsize=(10,5))
    fig.tight_layout()

    for i, attribute in enumerate(ATTRIBUTES):
        x = win_percentage.keys()
        y = [ win_percentage[key] for key in x ]
        x = list(map(str, x))
        ax[i].bar(x, y)
        ax[i].set_xlabel(attribute)

    plt.savefig(output, format=args.output.split('.')[-1])

def get_games_by_num_players(results):
    num = defaultdict(int)
    num_won = defaultdict(int)
    for result in results:
        num[result["numPlayers"]] += 1
        if result["place"] == 1:
            num_won[result["numPlayers"]] += 1

    ratios = {i: num_won[i] / num[i] for i in num}
    return num, num_won, ratios

def get_names(results):
    enemy_names = defaultdict(int)
    our_names = defaultdict(int)
    lost_names = defaultdict(int)
    for result in results:
        for enemy_name in result["enemyNames"]:
            enemy_names[enemy_name] += 1
        if result["ourName"] != '':
            our_names[result["ourName"]] += 1
        if result["place"] != 1:
            for active_enemy_name in result["endActiveEnemies"]:
                lost_names[active_enemy_name] += 1
        else:
            assert(len(result["endActiveEnemies"])) == 0
    return enemy_names, our_names, lost_names
    


def get_parser():
    parser = argparse.ArgumentParser(description="process logfiles and show statistics")
    parser.add_argument("files", metavar="file", help="json logfiles or directory structures", nargs='+')
    parser.add_argument("--all", "-a", help="print all available stats", action='store_true')
    parser.add_argument("--names", help="print name stats", action='store_true')
    parser.add_argument("--client", "-c", help="print client stats", action='store_true')
    parser.add_argument("--num", "-n", help="print number of players stats", action='store_true')
    parser.add_argument("--attributes", help="print attribute stats", action='store_true')
    parser.add_argument("--output", "-o", help="attribute diagram output file, can be jpeg, png, svg")
    return parser

if __name__ == '__main__':
    args = get_parser().parse_args()
    if not args.all and not args.names and not args.client and not args.num and not args.attributes:
        print("which stats do you want to have?")
        get_parser().print_usage()
        exit(1)

    results = get_results(args.files)

    print()

    if args.names or args.all:
        enemy_names, our_names, lost_names = get_names(results)

        print("we had the names", ", ".join(our_names))
        print(len(enemy_names), "names occured, lost against", len(lost_names), "=", '{:.1f}%'.format(100 * len(lost_names) / len(enemy_names)), "of names")
        print("total\twon\tratio\tname")
        for name in lost_names:
            print(enemy_names[name], enemy_names[name] - lost_names[name], "{:.1f}%".format((enemy_names[name] - lost_names[name]) / enemy_names[name] * 100), name, sep="\t")
        print()

    if args.num or args.all:
        print("total\twon\tratio\tplayers")
        total_games, won_games, win_ratio = get_games_by_num_players(results)
        for num_players in total_games:
            print(total_games[num_players], won_games[num_players], "{:.1f}%".format(win_ratio[num_players]*100), num_players, sep="\t")
        print()

    if args.attributes or args.all:
        total_games, won_games, win_ratio = get_attribute_stats(results)
        for attribute in total_games:
            print(attribute)
            print("total\twon\tratio\tvalue")
            for value in total_games[attribute]:
                print(total_games[attribute][value], won_games[attribute][value], "{:.1f}%".format(100*win_ratio[attribute][value]), value, sep="\t")
            print()
        if args.output is not None:
            create_attribute_diagram(results, args.output)
    if args.client or args.all:
        total_games, won_games, win_ratio = get_client_stats(results)
        print("total\twon\tratio\tclient")
        for client in total_games:
            print(total_games[client], won_games[client], "{:.1f}%".format(win_ratio[client]*100), client, sep="\t")
        print()

