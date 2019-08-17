
# FUll GUI for QSM

# Goal
The goal is to move to more Mouse driven navigation from the current pure OpenGL control based on key strokes.

# Tech option
This should use Nuklear on top of the current OpenGL pure 3D canvas

# Steps and issues:

- Convert all the key selections flags and size to a dialog box
    - The best will be like Gimp layout. To have external winfow to control all this independent of the window of OpenGL 3D rendering
    - The SizeVar object should be linked to a slider or value box. The current SizeVar there are (from [DisplayWorld->DisplaySettings](https://github.com/freddy33/qsm-go/blob/master/m3gl/m3gl.go) :
        - Line Width
        - Sphere Radius
        - FOV Angle
        - Eye Dist
    - All the above SizeVar are controlled by keystrokes and also GUI slider. The available keystrokes should be visible in the GUI.
    - Flags and Selector from [SpaceDrawingFilter->DisplaySettings](https://github.com/freddy33/qsm-go/blob/master/m3gl/m3drawing.go) :
        - Flag: Display Empty Nodes
        - Flag: Display Empty Connections
        - Positive Integer: Event Outgrowth Threshold
        - Positive Integer: Event Outgrowth Many Colors Threshold
        - Checkboxes: Event Colors Mask
    - The flags should be checkboxes, positive integer just a slider with max at 4, and mask using list of checkboxes

- Information Display window
    - A new window to display information about the world like shown in [Space->DisplayState](https://github.com/freddy33/qsm-go/blob/master/m3space/m3space.go)
    - Then under the window a selected object info box

- Find a way to identified objects under the mouse click:
    - First print out the actual coordinates of a mouse click in the 3D canvas
    - Use the inverse of the projection matrix to find the object: http://antongerdelan.net/opengl/raycasting.html
    - Google search "ow to find objects in OpenGL under a mouse click golang" provides good insights

- List the current visible objects based on the flags and filters
- Use the above list to find object in click ray and then display object characteristics in the info window


