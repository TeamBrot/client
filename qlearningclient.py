import asyncio
import websockets
import random
import json

#JSON to send for corresponding action
LEFT = '{"action":"turn_left"}'
RIGHT = '{"action":"turn_right"}'
SLOWER = '{"action":"slow_down"}'
FASTER = '{"action":"speed_up"}'
NOTHING = '{"action":"change_nothing"}'


async def connect():
    uri = "ws://localhost:8080/spe_ed"
    epsilon = 0.01
    q_table = init_q_table()
    last_state = None
    last_action = None
    #conect to websocket
    async with websockets.connect(uri) as websocket:
        while(True):
            # wait for data from server
            data = await websocket.recv()
            state = json.loads(data)
            # use data to calculate next move
            l = Logik(state, epsilon, q_table, last_state, last_action)
            epsilon, q_table, action = l.calculate_action()
            epsilon = get_new_epsilon(epsilon)
            # send next move back to server
            await websocket.send(action)
            last_state = state
            last_action = action

def get_new_epsilon(epsilon):
    epsilon = epsilon * 2
    if epsilon > 1:
        return 1
    else:
        return epsilon

def init_q_table():
    q_table = []
    try:
        file = open("q-table.txt", "r")
    except:
        file = open("q-table.txt", "w")
        file.close()
        file = open("q-table.txt", "r")
    content = []
    for line in file:
        content += [line.rstrip()]
    file.close()
    # epsilon = float(content[0])
    # content.remove(content[0])
    for q_val in content:
        q_val_list = [[], []]
        state = q_val.split("|")[0].split(",")
        values = q_val.split("|")[1].split(",")
        for l in state:
            q_val_list[0].append(l.split())
        q_val_list[1] = values
        for i in range(len(q_val_list[0])):
            for j in range(len(q_val_list[0][i])):
                q_val_list[0][i][j] = int(q_val_list[0][i][j])
        for i in range(5):
            q_val_list[1][i] = float(q_val_list[1][i])
        q_table.append(q_val_list)
    return q_table

def save_q_table(q_table):
    file = open("q-table.txt", "w")
    for entry in q_table:
        for state in range(len(entry[0])):
            for cell in range(len(entry[0][state])):
                file.write(str(entry[0][state][cell]))
                if cell != len(entry[0][state])-1:
                    file.write(" ")
            if state != len(entry[0])-1:
                file.write(",")
        file.write("|")
        for q_val in range(len(entry[1])):
            file.write(str(entry[1][q_val]))
            if q_val != len(entry[1])-1:
                file.write(",")
        file.write("\n")
    file.close()

class Logik:
    def __init__(self, state, epsilon, q_table, last_state, last_action):
        self.width = state["width"]
        self.height = state["height"]
        self.cells = state["cells"]
        self.players = state["players"]
        self.name = state["you"]
        self.running = state["running"]

        if last_state != None:
            self.last_width = last_state["width"]
            self.last_height = last_state["height"]
            self.last_cells = last_state["cells"]
            self.last_players = last_state["players"]
            self.last_name = last_state["you"]
            self.last_running = last_state["running"]

        self.last_state = last_state
        self.last_action = last_action


        self.actions = [LEFT, RIGHT, SLOWER, FASTER, NOTHING]

        self.gamma = 0.99

        self.epsilon = epsilon
        self.q_table = q_table

    def calculate_action(self):
        if self.last_state == None:
            return self.epsilon, self.q_table, random.choice(self.actions)
        q_index = self.update_q_table() #last state with last action
        random.seed()
        a = random.random()
        if a > self.epsilon:
            return self.epsilon, self.q_table, random.choice(self.actions) #currently chooses random only --> implement epsilon
        else:
            q_values = self.q_table[q_index][1]
            action_index = q_values.index(max(q_values))
            return self.epsilon, self.q_table, self.actions[action_index]

    def update_q_table(self):
        r_val = self.calculate_r_val()
        area_state = self.get_area_state(3, False, self.cells, self.players)
        last_area_state = self.get_area_state(3, True, self.last_cells, self.last_players)
        q_index = self.find_q_index(area_state)
        last_q_index = self.find_q_index(last_area_state)
        action_index = self.actions.index(self.last_action)
        q_neu = r_val + self.gamma * self.get_max_q_val(q_index)
        self.q_table[last_q_index][1][action_index] = q_neu
        if not self.running:
            save_q_table(self.q_table)
        return q_index

    def find_q_index(self, area_state):
        for i in self.q_table:
            print(i[0], area_state)
            if i[0] == area_state:
                print(i[0], area_state, self.q_table.index(i))
                return self.q_table.index(i)
        self.q_table.append([area_state, [0, 0, 0, 0, 0]])
        return len(self.q_table)-1

    def get_max_q_val(self, q_index):
        q_values = self.q_table[q_index][1]
        return max(q_values)

    def get_area_state(self, n, last_state, cells, players):
        #n = 2k-1
        position = [players[str(self.name)]["x"], players[str(self.name)]["y"]]
        area_start = [int(position[0]-(n-1)/2), int(position[1]-(n-1)/2)]
        area_state = []
        for i in range(n):
            area_state.append([])
            for j in range(n):
                if area_start[0]+j < 0 or area_start[1]+i < 0 or area_start[0]+j > self.height-1 or area_start[1]+i > self.width-1:
                    area_state[i] += [-2]   # Wall
                else:
                    cell = cells[area_start[1]+i][area_start[0]+j]
                    if cell != 0:
                        area_state[i] += [1]   # any player --> distinguish players???
                    else:
                        area_state[i] += [0]   # empty
        return area_state



    def calculate_r_val(self):
        if not self.running:
            return -100
        else:
            return 10


# useData uses the received game data to return the next actions
# def useData(data):
#     # load json data
#     d = json.loads(data)
#     print(d)
#     # extract information from json (examples)
#     deadline = d["deadline"]
#     you = d["you"]
#     own_direction = d["players"][str(you)]["direction"]
#     #...
#
#     #print("Your dir:" + d["players"][str(you)]["direction"])
#     #print("Other dir:" + d["players"][str(you+1)]["direction"])
#
#     #move diagonally
#     if(own_direction == "right"):
#         return RIGHT
#     elif(own_direction == "down"):
#         return LEFT
#
#     return NOTHING



#main
asyncio.get_event_loop().run_until_complete(connect())
asyncio.get_event_loop().run_forever()
print("Fertig")