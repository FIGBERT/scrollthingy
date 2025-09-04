import { Room, RoomEvent, Track } from "livekit-client";

const room = new Room();

export async function connect_to_room(url, token) {
  room
    .on(RoomEvent.TrackSubscribed, handleTrackSubscribed)
    .on(RoomEvent.TrackUnsubscribed, handleTrackUnsubscribed)
    .on(RoomEvent.Disconnected, handleDisconnect)
  await room.connect(url, token);
}

export async function send_scroll(delta) {
  await room.localParticipant.sendText(delta, { topic: "scroll-updates" });
}

function handleTrackSubscribed(track, _publication, _participant) {
  if (track.kind === Track.Kind.Video || track.kind === Track.Kind.Audio) {
    // attach it to a new HTMLVideoElement or HTMLAudioElement
    const element = track.attach();
    const parent = document.getElementById("livekit");
    parent.appendChild(element);
  }
}

function handleTrackUnsubscribed(track, _publication, _participant) {
  // remove tracks from all attached elements
  track.detach();
}

function handleDisconnect() {
  console.log("disconnected from room");
}
