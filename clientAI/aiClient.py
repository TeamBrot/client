import websocket
from information import game
from information import start_state, transform
import numpy as np
import tensorflow as tf
from tensorflow import keras
from tensorflow.keras import layers


# Variables for connection
connected = False
ws = None
actions = {
    0 : '{"action":"turn_left"}',
    1 : '{"action":"turn_right"}',
    2 : '{"action":"slow_down"}',
    3 : '{"action":"speed_up"}',
    4 : '{"action":"change_nothing"}'
    }

# Variables for Keras
# Configuration paramaters for the whole setup
seed = 42
gamma = 0.99                    # Discount factor for past rewards
epsilon = 1.0                   # Epsilon greedy parameter
epsilon_min = 0.1               # Minimum epsilon greedy parameter
epsilon_max = 1.0               # Maximum epsilon greedy parameter
epsilon_interval = (
    epsilon_max - epsilon_min
)                               # Rate at which to reduce chance of random action being taken
batch_size = 32                 # Size of batch taken from replay buffer
max_steps_per_episode = 10000

state = None
episode_reward = 0
first_msg = True
action = 0

# Called when websocket connection is closed
def on_close(ws):
    global connected
    if(not connected):
        return
    connected = False
    print("close")

def on_error(ws, error):
    ... 

# called when the websocket recieves a message
def on_message(ws, message):
    #print(message)
    ws.send(ai(message))

# called when the websocket is opened
def on_open(ws):
    global connected
    connected = True
    print("opened")

    global first_msg
    first_msg = True

    global state
    global episode_reward
    state = np.array(start_state())
    episode_reward = 0

def connect(uri = "ws://localhost:8080/spe_ed"):
    """Connect to URI"""
    # if(ws):
    #     ws.close()
    global ws
    websocket.enableTrace(True)
    ws = websocket.WebSocketApp(uri,
                        on_open = on_open,
                        on_message = on_message,
                        on_error = on_error,
                        on_close = on_close)
    ws.run_forever()


def ai(data):
    gamestate = game(data)
    #return actions["nothing"]

    global frame_count
    global action
    global state
    global episode_reward
    global epsilon
    global epsilon_interval
    global epsilon_greedy_frames

    # Apply the sampled action in our environment
    state_next = transform(gamestate)
    reward = 1
    done = False

    state_next = np.array(state_next)

    episode_reward += reward

    # Save actions and states in replay buffer
    action_history.append(action)
    state_history.append(state)
    state_next_history.append(state_next)
    done_history.append(done)
    rewards_history.append(reward)
    state = state_next

    # Update every fourth frame and once batch size is over 32
    if frame_count % update_after_actions == 0 and len(done_history) > batch_size:
        # Get indices of samples for replay buffers
        indices = np.random.choice(range(len(done_history)), size=batch_size)

        # Using list comprehension to sample from replay buffer
        state_sample = np.array([state_history[i] for i in indices])
        state_next_sample = np.array([state_next_history[i] for i in indices])
        rewards_sample = [rewards_history[i] for i in indices]
        action_sample = [action_history[i] for i in indices]
        done_sample = tf.convert_to_tensor(
            [float(done_history[i]) for i in indices]
        )

        # Build the updated Q-values for the sampled future states
        # Use the target model for stability
        future_rewards = model_target.predict(state_next_sample)
        # Q value = reward + discount factor * expected future reward
        updated_q_values = rewards_sample + gamma * tf.reduce_max(
            future_rewards, axis=1
        )

        # If final frame set the last value to -1
        updated_q_values = updated_q_values * (1 - done_sample) - done_sample

        # Create a mask so we only calculate loss on the updated Q-values
        masks = tf.one_hot(action_sample, num_actions)

        with tf.GradientTape() as tape:
            # Train the model on the states and updated Q-values
            q_values = model(state_sample)

            # Apply the masks to the Q-values to get the Q-value for action taken
            q_action = tf.reduce_sum(tf.multiply(q_values, masks), axis=1)
            # Calculate loss between new Q-value and old Q-value
            loss = loss_function(updated_q_values, q_action)

        # Backpropagation
        grads = tape.gradient(loss, model.trainable_variables)
        optimizer.apply_gradients(zip(grads, model.trainable_variables))

    if frame_count % update_target_network == 0:
        # update the the target network with new weights
        model_target.set_weights(model.get_weights())
        # Log details
        template = "running reward: {:.2f} at episode {}, frame count {}"
        print(template.format(running_reward, episode_count, frame_count))

    # Limit the state and reward history
    if len(rewards_history) > max_memory_length:
        del rewards_history[:1]
        del state_history[:1]
        del state_next_history[:1]
        del action_history[:1]
        del done_history[:1]

    if done:
        ws.close


    frame_count += 1
    # Use epsilon-greedy for exploration
    if frame_count < epsilon_random_frames or epsilon > np.random.rand(1)[0]:
        # Take random action
        action = np.random.choice(num_actions)
    else:
        # Predict action Q-values
        # From environment state
        state_tensor = tf.convert_to_tensor(state)
        state_tensor = tf.expand_dims(state_tensor, 0)
        action_probs = model(state_tensor, training=False)
        # Take best action
        action = tf.argmax(action_probs[0]).numpy()

    # Decay probability of taking random action
    epsilon -= epsilon_interval / epsilon_greedy_frames
    epsilon = max(epsilon, epsilon_min)
    return actions[action]


# Creating the AI Model

# number of possible actions
num_actions = 5


def create_q_model():
    # Network defined by the Deepmind paper
    inputs = layers.Input(shape=(40, 40, 3,))

    # Convolutions on the frames on the screen
    layer1 = layers.Conv2D(32, 8, strides=4, activation="relu")(inputs)
    layer2 = layers.Conv2D(64, 4, strides=2, activation="relu")(layer1)
    layer3 = layers.Conv2D(64, 3, strides=1, activation="relu")(layer2)

    layer4 = layers.Flatten()(layer3)

    layer5 = layers.Dense(512, activation="relu")(layer4)
    action = layers.Dense(num_actions, activation="linear")(layer5)

    return keras.Model(inputs=inputs, outputs=action)

# The first model makes the predictions for Q-values which are used to
# make a action.
model = create_q_model()
# Build a target model for the prediction of future rewards.
# The weights of a target model get updated every 10000 steps thus when the
# loss between the Q-values is calculated the target Q-value is stable.
model_target = create_q_model()

# In the Deepmind paper they use RMSProp however then Adam optimizer
# improves training time
optimizer = keras.optimizers.Adam(learning_rate=0.00025, clipnorm=1.0)

# Experience replay buffers
action_history = []
state_history = []
state_next_history = []
rewards_history = []
done_history = []
episode_reward_history = []
running_reward = 0
episode_count = 0
frame_count = 0
# Number of frames to take random action and observe output
epsilon_random_frames = 50000
# Number of frames for exploration
epsilon_greedy_frames = 1000000.0
# Maximum replay length
# Note: The Deepmind paper suggests 1000000 however this causes memory issues
max_memory_length = 100000
# Train the model after 4 actions
update_after_actions = 4
# How often to update the target network
update_target_network = 10000
# Using huber loss for stability
loss_function = keras.losses.Huber()

if __name__ == "__main__":
    while(True):
        connect()
        while(connected):
            ...


