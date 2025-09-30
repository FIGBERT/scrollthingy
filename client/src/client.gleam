import gleam/dynamic/decode
import gleam/http/response
import gleam/int
import gleam/json
import gleam/result
import gleam/string
import lustre
import lustre/attribute as attr
import lustre/effect.{type Effect}
import lustre/element.{type Element}
import plinth/browser/event.{type Event}
import plinth/browser/window
import rsvp
import sketch.{type StyleSheet}
import sketch/css
import sketch/css/length
import sketch/lustre as skls
import sketch/lustre/element/html

const server = "http://localhost:8080"

type Model {
  Intro
  Leader
  Line(position: Int, size: Int)
}

type Msg {
  Wheel(delta: Int)
  ConnectTo(url: String, token: String)
  RoomUpdate(idx: Int, total: Int)
  Err(String)
}

@external(javascript, "./scroll.mjs", "delta_from_event")
fn delta_from_event(evt: Event(a)) -> Int

@external(javascript, "./livekit.mjs", "connect_to_room")
fn connect_to_room(
  url: String,
  token: String,
  dispatch: fn(String) -> Nil,
) -> Nil

@external(javascript, "./livekit.mjs", "send_scroll")
fn send_scroll(delta: Int) -> Nil

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
  |> result.map_error(fn(_) { Err("Unable to connect to the room.") })
  |> result.unwrap_both
}

fn connect_effect(url: String, token: String) -> Effect(Msg) {
  effect.from(fn(dispatch) {
    let payload_decoder = {
      use idx <- decode.field("index", decode.int)
      use total <- decode.field("total", decode.int)
      decode.success(RoomUpdate(idx:, total:))
    }

    let js_issue_msgs = fn(update: String) -> Nil {
      let result = json.parse(update, payload_decoder)
      case result {
        Ok(msg) -> dispatch(msg)
        Error(_) -> dispatch(Err("Unable to parse room update."))
      }
    }

    connect_to_room(url, token, js_issue_msgs)
  })
}

fn scroll_effect(delta: Int) -> Effect(Msg) {
  effect.from(fn(_dispatch) { send_scroll(delta) })
}

fn init(_args: Nil) -> #(Model, Effect(Msg)) {
  #(Intro, effect.batch([listen_for_scroll(), get_token()]))
}

fn update(model: Model, msg: Msg) -> #(Model, Effect(Msg)) {
  case msg {
    Wheel(delta) -> #(model, scroll_effect(delta))
    ConnectTo(url, token) -> #(model, connect_effect(url, token))

    RoomUpdate(idx, total) -> {
      case idx {
        0 -> #(Leader, effect.none())
        _ -> #(Line(idx, total - 1), effect.none())
      }
    }

    Err(_) -> #(model, effect.none())
  }
}

fn view(model: Model, styles: StyleSheet) -> Element(Msg) {
  use <- skls.render(stylesheet: styles, in: [skls.node()])

  html.div_([], [
    html.video(
      css.class([css.width(length.percent(100))]),
      [attr.id("livekit")],
      [],
    ),
    html.p(
      css.class([
        css.position("absolute"),
        css.inset_inline_start("5ch"),
        css.inset_block_start("3ch"),
        css.color("red"),
        css.font_weight("bold"),
      ]),
      [],
      [
        html.text(case model {
          Intro -> "intro"
          Leader -> "leader"
          Line(pos, total) ->
            int.to_string(pos) <> "/" <> int.to_string(total) <> " in-line"
        }),
      ],
    ),
  ])
}

pub fn main() -> Nil {
  let assert Ok(styles) =
    skls.construct(fn(styles) {
      styles
      |> sketch.global(
        css.global("html, body", [
          css.margin(length.px(0)),
          css.height(length.percent(100)),
          css.overflow("hidden"),
        ]),
      )
    })

  let assert Ok(_) =
    lustre.application(init, update, view(_, styles))
    |> lustre.start("body", Nil)
  Nil
}
