import json
import os
import sys
import argparse
from datetime import datetime

# Define paths relative to workspace root
LOG_DIR = ".agent"
LOG_FILE = os.path.join(LOG_DIR, "state_log.json")

def load_log():
    if not os.path.exists(LOG_DIR):
        os.makedirs(LOG_DIR)
        
    if os.path.exists(LOG_FILE):
        try:
            with open(LOG_FILE, 'r') as f:
                return json.load(f)
        except json.JSONDecodeError:
            pass # File is empty or corrupt, fallback to template
            
    return {
        "current_objective": "Initializing persistent state monitoring.",
        "why": "Preventing AI context loss during extended development sessions.",
        "active_problems": [],
        "completed_milestones": [],
        "next_step": "Awaiting initial prompt.",
        "last_updated": ""
    }

def save_log(log_data):
    # Enforce rolling limits to prevent token bloat
    log_data["active_problems"] = log_data["active_problems"][-5:]
    log_data["completed_milestones"] = log_data["completed_milestones"][-5:]
    log_data["last_updated"] = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    
    with open(LOG_FILE, 'w') as f:
        json.dump(log_data, f, indent=2)

def main():
    parser = argparse.ArgumentParser(description="Update AntiGravity persistent memory state.")
    parser.add_argument('--objective', type=str, help='The core current programming objective.')
    parser.add_argument('--why', type=str, help='The architectural reason or background motivation.')
    parser.add_argument('--problem', type=str, help='An active error message, block, or bug encountered.')
    parser.add_argument('--clear_problems', action='store_true', help='Clear out resolved issues.')
    parser.add_argument('--milestone', type=str, help='A successful fix or completed block of logic.')
    parser.add_argument('--next', type=str, help='The immediate next step the AI needs to take.')

    args = parser.parse_args()
    log_data = load_log()

    if args.objective:
        log_data["current_objective"] = args.objective
    if args.why:
        log_data["why"] = args.why
    if args.next:
        log_data["next_step"] = args.next
        
    if args.clear_problems:
        log_data["active_problems"] = []
    elif args.problem:
        log_data["active_problems"].append(args.problem)
        
    if args.milestone:
        log_data["completed_milestones"].append(args.milestone)

    save_log(log_data)
    print(f"Successfully updated persistent state at {log_data['last_updated']}.")

if __name__ == "__main__":
    main()