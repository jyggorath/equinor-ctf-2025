import socket
import threading
import secrets
import time
from time import sleep

# Game settings
ANSWER_ROTATION_INTERVAL = 60  # Rotate answers every x seconds
MAX_CONNECTIONS = 25
PORT = 5555

emoji_list = [
    "😊", "❤️", "⭐", "🔥", "🌈", 
    "🎉", "🎈", "🌟", "🍀", "🍕", 
    "🎂", "🌍", "🚀", "💎", "🎵", 
    "🐾", "🌻", "🦄", "⚡", "🌙"
]
emoji_list_2 = [
    "😄", "💖", "🌊", "🍉", "🌺", 
    "🥳", "🍔", "🦋", "🍂", "🏆", 
    "🥇", "🌼", "🌈", "🍦", "🐉", 
    "🌴", "🧩", "🎤", "🌌", "🧙‍♂️"
]

# Answers for each level
level1_answers = list(range(10))                      # Numbers 0-9
level2_answers = [chr(i) for i in range(65, 91)]      # Letters A-Z
level3_answers = emoji_list                           # Emojis
level4_answers = emoji_list_2                         # Emojis
level5_answers = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!"

# Current answers
current_answers = {
    "level1": None,
    "level2": None,
    "level3": None,
    "level4": None,
    "level5": None
}

def eval_guess(guess, correct_guess, sleep_seconds=0.5):
    sleep(sleep_seconds)
    print (guess, correct_guess)
    return guess != correct_guess
# Rotate answers for each level
def rotate_answers():
    while True:
        current_answers["level1"] = secrets.choice(level1_answers)
        current_answers["level2"] = secrets.choice(level2_answers)
        current_answers["level3"] = secrets.choice(level3_answers)
        current_answers["level4"] = secrets.choice(level4_answers)
        current_answers["level5"] = secrets.choice(level5_answers)
        time.sleep(ANSWER_ROTATION_INTERVAL)

# Handle individual client connections
def handle_client(client_socket):
    try:
        client_socket.send(b"Welcome to the Guessing Game!\n1. Start Game\n2. Exit\nChoose an option: ")
        option = client_socket.recv(1024).strip().decode()

        if option == "2":
            client_socket.send(b"Goodbye!\n")
            client_socket.close()
            return

        # Start Level 1
        client_socket.send(b"Level 1. Guess a number between 0-9:\n")
        guess = client_socket.recv(1024).strip().decode()
        # print (guess, str(current_answers["level1"]))
        if eval_guess(guess, str(current_answers["level1"])):
            client_socket.send(b"Wrong guess. Game over.\n")
            client_socket.close()
            return

        # Start Level 2
        client_socket.send(b"Level 2. Guess a letter between A-Z:\n")
        guess = client_socket.recv(1024).strip().decode()
        if eval_guess(guess, current_answers["level2"]):
            client_socket.send(b"Wrong guess. Game over.\n")
            client_socket.close()
            return

        # Start Level 3
        client_socket.send(f"Level 3. Guess an emoji ({' '.join(emoji_list)}):\n".encode('utf-8'))
        guess = client_socket.recv(1024).strip().decode()
        if eval_guess(guess, current_answers["level3"]):
            client_socket.send(b"Wrong guess. Game over.\n")
            client_socket.close()
            return
        
        # Start Level 4
        client_socket.send(f"Level 4. Guess an emoji ({' '.join(emoji_list_2)}):\n".encode('utf-8'))
        guess = client_socket.recv(1024).strip().decode()
        if eval_guess(guess, current_answers["level4"]):
            client_socket.send(b"Wrong guess. Game over.\n")
            client_socket.close()
            return
        
        # Start Level 5
        client_socket.send(f"Level 5. Guess a letter ({level5_answers}):\n".encode())
        guess = client_socket.recv(1024).strip().decode()
        if eval_guess(guess, current_answers["level5"]):
            client_socket.send(b"Wrong guess. Game over.\n")
            client_socket.close()
            return

        # Player has won
        flag = open("flag.txt", "r").read().strip()
        client_socket.send(b"Congratulations! You won the game! FLAG:" + flag.encode() + b"\n")
    except Exception as e:
        print(f"An error occurred: {e}")
    finally:
        client_socket.close()

# Main server function
def start_server(host="0.0.0.0", port=PORT):
    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.bind((host, port))
    server.listen(MAX_CONNECTIONS)
    print(f"Server listening on {host}:{port}")

    # Start rotating answers in a separate thread
    answer_thread = threading.Thread(target=rotate_answers, daemon=True)
    answer_thread.start()

    # Handle incoming connections
    while True:
        client_socket, addr = server.accept()
        print(f"Accepted connection from {addr}")
        client_handler = threading.Thread(target=handle_client, args=(client_socket,))
        client_handler.start()

if __name__ == "__main__":
    start_server()
