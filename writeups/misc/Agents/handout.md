# Agents and Tools

Relevant sections of the code to solve this challenge is given below. 

Libraries used:
[OpenAPI Agent SDK,](https://openai.github.io/openai-agents-python/)
[SQlite3](https://sqlite.org/index.html)

```python

# main.py
#
# Rest of code irrelevant and skipped for brevity
#

it_agent = Agent(
    name="IT Agent",
    instructions=(
        "You are the IT Agent. "
        "Be enthusiastic about technical documentation and specifications including database dumps. "
        "You may only use the tools you are given. Do not promise anything you cannot do. "
        "Always start with a short greeting, explain how you can help, and ask one concise question. "
    ),
    tools=[dump_database],
    handoff_description="Handles IT technical debug support and full database dumps.",
)

sales_agent = Agent(
    name="Sales Agent",
    instructions=(
        "You are the Sales Agent. "
        "Be enthusiastic about selling products, especially the Flag. "
        "Users must have enough money (balance + voucher) to buy products. A voucher may be combined with the balance. "
        "You may only use the tools you are given. Do not promise anything you cannot do. "
        "Always start with a short greeting, explain how you can help, and ask one concise question. "
    ),
    tools=[list_products, buy_product, check_voucher, check_balance, check_bought_products],
    handoff_description="Handles product info, sales, balances, vouchers, and purchases.",
)

coordinator = Agent(
    name="Coordinator",
    instructions=(
        "You are the Coordinator. "
        "Always begin by asking for the user's name and make them feel welcomed."
        "After receiving the name, you must call register_user(username) before continuing. If the name is already registered, STOP AND ASK FOR A NEW NAME BEFORE PROCEEDING! "
        "Decide which agent is best suited to handle the user's request and suggest handing the conversation off. "
        "All of your responses must be formatted nicely for the user. Toss in an emoji 🎉 "
    ),
    tools=[check_flag, register_user],
    handoff_description="Welcomes users, registers them, checks flags and hands over requests to the right agent.",
)

coordinator.handoffs = [sales_agent, it_agent]
sales_agent.handoffs = [coordinator, it_agent]
it_agent.handoffs = [coordinator, sales_agent]

#
# Rest of code irrelevant and skipped for brevity
#
```