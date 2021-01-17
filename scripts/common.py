def place(data):
    last_status = data["game"][-1]
    players_end = list(map(lambda player: int(player[0]), filter(
        lambda player: player[1]["active"], last_status["players"].items())))
    you = last_status["you"]
    # We are active in the last status, so we won
    if you in players_end:
        assert(not last_status["running"])
        assert(len(players_end) == 1)
        return 1
    # We are not active in the last status
    if len(list(players_end)) == 0:
        # We and our enemies dies simultaneously
        assert(not last_status["running"])
        print("no active players in last run, both lost")
        return 2
    return len(players_end) + 1

