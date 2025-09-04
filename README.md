# The Scroll Pun: Development Scratchpad

## Current Status
The device has been constructed. Now we're rapidly iterating through the
software. As of just now, it can seperately:

- Track user scroll on the client and send to the server
- Broadcast (server) and receive (client) video using LiveKit+Pion

The latter feature relies on a lot of Javascript, which exists outside
the Gleam/Lustre lifecycle in a way that makes me uncomfortable. I
either need to bring it into the fold, or move away from that stack. I
like the stack so will probably do the former.

## Features To Be Developed
Participants need to be managed more closely. Right now they are
semi-ephemeral: token are issued with random IDs, but those are not
stored anywhere and live only in LiveKit. They should instead be listed
in an ordered queue. Though I suppose server-side, we only need to store
the scroll offset of the active user.

Frank and I have been planning on putting users into a line to scroll
from the outset of the project, but David had the great idea just now to
allow them to scroll through the list of folks waiting. Scroll-ception.

Having written this down I think I can see it coming together in my
mind. I'll be back soon.
