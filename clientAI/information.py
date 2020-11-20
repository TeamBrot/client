import json
class player(object):
    """Creates a player object from the given json"""
    def __init__(self, data):
        self.x = data["x"]
        self.y = data["y"]
        self.direction = data["direction"]
        self.speed = data["speed"]
        self.active = data["active"]
        self.name = data["name"]

class game(object):
    """Creates an info object with all the information sent by the server"""
    def __init__(self, data):
        try:
            d = json.loads(data)
            self.width = d["width"]
            self.height = d["height"]
            self.cells = d["cells"]
            self.players = d["players"]
            self.you = d["you"]
            self.running = d["running"]
            self.deadline = d["deadline"]

            self.player_you = player(self.players[str(self.you)])
        except:
            print("An error occured trying to decode the server message")

def start_state(width=40, height=40):
    observation = []
    for h in range(height):
        row =[]
        for w in range(width):
            row.append([0,0,0])  
        observation.append(row)
    return observation

def transform(g):
    # transforming the cells into observation
    observation = []
    for h in range(g.height):
        row =[]
        for w in range(g.width):
            field = [0,0,0]
            p = g.cells[h][w]
            if(p==0):
                field[0] = 1
                row.append(field)
            elif(p==1):
                field[1] = 1
                row.append(field)
            elif(p==2):
                field[2] = 1
                row.append(field)
            else:
                row.append(field)    
        observation.append(row)
    return observation
