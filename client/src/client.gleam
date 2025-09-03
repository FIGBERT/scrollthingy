import gleam/http/response
import gleam/int
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
  Token(str: String)

  Noop
}

@external(javascript, "./scroll.mjs", "delta_from_event")
fn delta_from_event(evt: Event(a)) -> Int

@external(javascript, "./livekit.mjs", "connect_to_room")
fn connect_to_room(token: String) -> Nil

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
  case resp {
    Ok(resp) -> Token(resp.body)
    Error(_) -> Noop
  }
}

fn connect_effect(token: String) -> Effect(Msg) {
  effect.from(fn(dispatch) {
    connect_to_room(token)
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
    Token(value) -> #(model, connect_effect(value))
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
