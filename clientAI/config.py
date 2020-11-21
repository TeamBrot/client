seed = 42
gamma = 0.99                    # Discount factor for past rewards
epsilon = 1.0                   # Epsilon greedy parameter
epsilon_min = 0.1               # Minimum epsilon greedy parameter
epsilon_max = 1.0 

batch_size = 32                 # Size of batch taken from replay buffer
max_steps_per_episode = 10000

# optimizer
learn_rate = 0.00025
clipnorm=1.0

# training
max_episodes = 100

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