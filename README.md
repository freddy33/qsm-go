# Quantum State Machine in Golang

# Principles
- No floating points. Quantum is the name, so *integers* should be able to express the Universe!
- *Space* is a graph with each *Node* being connected to *three* and only three other nodes. This is since: "Three shall be the number thou shalt count, and the number of the counting shall be three."
- Each space node is located in 3D space using a `Point(X,Y,Z int64)` for display purposes.
- *Time* is ticking... As an unsigned integer of course.
- Space *events* are identified and are happening on a node at a time.
- Event *outgrowth* are developing through the graph of nodes.

# Type Definitions

- *Point*: a location in integer cartesian space
- *Node*: is in a certain point connected to 3 other nodes
- *TickTime*: the unsigned int telling time
- *Event*: An ID at a node at a time
- *EventOutgrowth*: The collection of nodes affected by a certain event at a certain distance (Distance = Current Time - Event Tick Time)

# Demo

Here is the OpenGL output for 4 events in pyramid shape after 26 steps:
![](https://github.com/freddy33/qsm-go/raw/master/docs/screenshot1.png)
