Elevator Project
================

[![Build Status](https://build.shapingideas.fyi/job/thefuturezebras/job/project-thefuturezebras/job/master/2/badge/icon)](https://build.shapingideas.fyi/job/thefuturezebras/job/project-thefuturezebras/job/master/2/)

Folder structure
----------------
| Folder | Description |
|--------|-------------|
|pkg     | Contains independent modules that can be used in other projects without any modifications. Modules in this folder can not depend on internal project modules. |
| internal | Contains internal modules. It is not possible to import these modules from external projects.
| scripts | Scripts used to launch, test and build the code |


Summary
-------
In this project we have created software for controlling `n` elevators working in parallel across `m` floors.


Main requirements
-----------------
The elevators should behave after these requirements.

No orders are lost
 - Once the light on an hall call button (buttons for calling an elevator to that floor; top 6 buttons on the control panel) is turned on, an elevator should arrive at that floor
 - Similarly for a cab call (for telling the elevator what floor you want to exit at; front 4 buttons on the control panel), but only the elevator at that specific workspace should take the order
 - This means handling network packet loss, losing network connection entirely, software that crashes, and losing power - both to the elevator motor and the machine that controls the elevator
 - If the elevator is disconnected from the network, it should still serve all the currently active orders (ie. whatever lights are showing)

Multiple elevators should be more efficient than one
 - The orders should be distributed across the elevators in a reasonable way
 - You are free to choose and design your own "cost function" of some sort: Minimal movement, minimal waiting time, etc.
 - The project is not about creating the "best" or "optimal" distribution of orders. It only has to be clear that the elevators are cooperating and communicating.
 
An individual elevator should behave sensibly and efficiently
 - No stopping at every floor "just to be safe"
 - The hall "call upward" and "call downward" buttons should behave differently
The lights and buttons should function as expected
 - The hall call buttons on all workspaces should let you summon an elevator
 - The lights on the hall buttons should show the same thing on all workspaces
 - The cab button lights should not be shared between elevators
 - The cab and hall button lights should turn on as soon as is reasonable after the button has been pressed
 - The cab and hall button lights should turn off when the corresponding order has been serviced
 - The "door open" lamp should be used as a substitute for an actual door, and as such should not be switched on while the elevator is moving

 
Start with `1 <= n <= 3` elevators, and `m == 4` floors. Try to avoid hard-coding these values: You should be able to add a fourth elevator with no extra configuration, or change the number of floors with minimal configuration. You do, however, not need to test for `n > 3` and `m != 4`.

   
Permitted assumptions
---------------------

The following assumptions will always be true during testing:
 - At least one elevator is always working normally
 - No multiple simultaneous errors: Only one error happens at a time, but the system must still return to a fail-safe state after this error
 - No network partitioning: There will never be a situation where there are multiple sets of two or more elevators with no connection between them
