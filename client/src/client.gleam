import gleam/http/response
import gleam/int
import gleam/result
import gleam/string
import lustre
import lustre/effect.{type Effect}
import lustre/element.{type Element}
import lustre/element/html
import plinth/browser/event.{type Event}
import plinth/browser/window
import rsvp

const server = "http://localhost:8080"

type Model =
  Int

type Msg {
  Wheel(delta: Int)
  ConnectTo(url: String, token: String)

  Noop
}

@external(javascript, "./scroll.mjs", "delta_from_event")
fn delta_from_event(evt: Event(a)) -> Int

@external(javascript, "./livekit.mjs", "connect_to_room")
fn connect_to_room(url: String, token: String) -> Nil

fn listen_for_scroll() -> Effect(Msg) {
  effect.from(fn(dispatch) {
    window.add_event_listener("wheel", fn(evt) {
      dispatch(Wheel(delta_from_event(evt)))
    })
  })
}

fn get_token() -> Effect(Msg) {
  let url = server <> "/token"
  rsvp.get(url, rsvp.expect_ok_response(process_token_response))
}

fn process_token_response(
  resp: Result(response.Response(String), rsvp.Error),
) -> Msg {
  resp
  |> result.map_error(fn(_) { Nil })
  |> result.try(fn(resp) { string.split_once(resp.body, on: "\n") })
  |> result.map(fn(tuple) { ConnectTo(url: tuple.0, token: tuple.1) })
  |> result.map_error(fn(_) { Noop })
  |> result.unwrap_both
}

fn connect_effect(url: String, token: String) -> Effect(Msg) {
  effect.from(fn(dispatch) {
    connect_to_room(url, token)
    dispatch(Noop)
  })
}

fn init(_args: Nil) -> #(Model, Effect(Msg)) {
  #(0, effect.batch([listen_for_scroll(), get_token()]))
}

fn update(model: Model, msg: Msg) -> #(Model, Effect(Msg)) {
  case msg {
    Wheel(delta) -> {
      let model = case delta {
        less if less < 0 -> model - 1
        more if more > 0 -> model + 1
        _ -> model
      }
      #(model, effect.none())
    }
    ConnectTo(url, token) -> #(model, connect_effect(url, token))
    Noop -> #(model, effect.none())
  }
}

fn view(model: Model) -> Element(Msg) {
  html.p([], [html.text(int.to_string(model))])
}

pub fn main() -> Nil {
  let assert Ok(_) =
    lustre.application(init, update, view)
    |> lustre.start("#lustre", Nil)
  Nil
}
