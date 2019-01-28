# Exercise 3: Uplink Established

This exercise is part of of a three-stage process:
 - This first exercise is to make you more familiar with using TCP and UDP for communication between processes running on different machines. Do not think too much about code quality here, as the main goal is familiarization with the protocols.
 - Exercise 4 will have you consider the things you have learned about these two protocols, and implement a network module that you can use in the project. The communication that is necessary for your design should reflect your choice of protocol. This is when you should start thinking more about code quality, because ...
 - Finally, you should be able to use this module as part of your project. Since predicting the future is notoriously difficult, you may find you need to change some functionality. But if the module has well-defined boundaries (a set of functions, communication channels, or others), you *should* be able to swap out the entire thing if necessary!


Note:
 - You are free to choose any language. Using the same language on the network exercises and the project is recommended, but not required. If you are still in the process of deciding, use this exercise as a language case study.
 - Exactly how you do communication for the project is up to you, so if you want to venture out into the land of libraries you should make sure that the library satisfies all your requirements. You should also check the license.


## Parts
- [Part 1: UDP](./Part1/README.md) 
- [Part 2: TCP](./Part2/README.md) 

## Approval
When asking for approval, make sure to prepare the following.
- Part 1:
    - The code for the UDP module.
    - Demonstrate that you can send an UDP message to the server and read the looped back message.
- Part 2:
    - The code for the TCP module.
    - Demonstrate that you can make the tcp server connect to your TCP server.
