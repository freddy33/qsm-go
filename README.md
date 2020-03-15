# Quantum State Machine in Golang

# What is QSM?
Also called Quantum Cellular Automaton
[Here](https://docs.google.com/presentation/d/1obaWWSiWTeHr3bYDxal9hmJ3GWxsTIpDBZtQgdUjxIk/edit?usp=sharing) is a presentation about the project.

# Principles
- No floating points. Quantum is the name, so *integers* should be able to express the Universe!
- *Space* is a graph with each *Node* being connected to *three* and only three other nodes. This is since: "Three shall be the number thou shalt count, and the number of the counting shall be three."
- Each space node is located in 3D space using a `Point(X,Y,Z int64)` for display purposes.
- *Time* is ticking... As an unsigned integer of course.
- Space *events* are identified and are happening on a node at a time.
- Event *outgrowth* are developing through the graph of nodes.

# Windows Installation
- Install WSL (Windows Subsystem for Linux) with Ubuntu ( info from https://www.howtogeek.com/249966/how-to-install-and-use-the-linux-bash-shell-on-windows-10/ ):
  - In PowerShell as Administrator run: `Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Windows-Subsystem-Linux`
  - Restart and in Microsoft Store add Ubuntu
- Run bash and test you have the needed applications:
  - sudo apt install jq
  - sudo apt install git
- Download and install golang: https://golang.org/dl/ NOTE: By default it's installed under `C:\Go` good for this
- Download and Install MinGW-W64: https://sourceforge.net/projects/mingw-w64/ IMPORTANT: Make sure to choose x86_64 architecture, and install under `C:\tools` it\'ll be easier.
  - Add the `mingw64/bin` folder in the global PATH windows env variable.


# Type Definitions

- *Point*: a location in integer cartesian space
- *Node*: is in a certain point connected to 3 other nodes
- *DistTime*: the unsigned int telling time
- *Event*: An ID at a node at a time
- *EventOutgrowth*: The collection of nodes affected by a certain event at a certain distance (Distance = Current Time - Event Tick Time)

# Demo

Here is the OpenGL output for 4 events in pyramid shape after 26 steps:
![](https://github.com/freddy33/qsm-go/raw/master/docs/screenshot1.png)
