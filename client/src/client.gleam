import gleam/dynamic/decode
import gleam/http/response
import gleam/int
import gleam/json
import gleam/option.{type Option, None, Some}
import gleam/result
import gleam/string
import lustre
import lustre/attribute as attr
import lustre/effect.{type Effect}
import lustre/element.{type Element}
import lustre/event as evt
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
  Intro(Option(Status))
  Game(Status)
}

type Status {
  Leader
  Line(position: Int, size: Int)
}

type Msg {
  Join
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
  #(Intro(None), effect.batch([listen_for_scroll(), get_token()]))
}

fn update(model: Model, msg: Msg) -> #(Model, Effect(Msg)) {
  case msg {
    Join -> {
      case model {
        Intro(Some(status)) -> #(Game(status), effect.none())
        _ -> #(Intro(None), effect.none())
      }
    }
    Wheel(delta) -> #(model, scroll_effect(delta))
    ConnectTo(url, token) -> #(model, connect_effect(url, token))

    RoomUpdate(idx, total) -> {
      let status = case idx {
        0 -> Leader
        _ -> Line(idx, total - 1)
      }
      case model {
        Game(_) -> #(Game(status), effect.none())
        Intro(_) -> #(Intro(Some(status)), effect.none())
      }
    }

    Err(_) -> #(model, effect.none())
  }
}

fn view(model: Model, styles: StyleSheet) -> Element(Msg) {
  use <- skls.render(stylesheet: styles, in: [skls.node()])

  let vid_class =
    css.class([
      css.width(length.percent(100)),
      css.height(length.percent(100)),
      css.position("absolute"),
      css.object_fit("cover"),
      ..case model {
        Intro(_) -> [css.filter("grayscale(1) brightness(0.4)")]
        Game(Line(_, _)) -> [css.filter("grayscale(1) brightness(0.4)")]
        _ -> []
      }
    ])

  let game_ui = case model {
    Intro(_) -> {
      html.div(
        css.class([
          css.position("absolute"),
          css.display("grid"),
          css.place_items("center"),
          css.width(length.percent(100)),
          css.height(length.percent(100)),
          css.text_align("center"),
        ]),
        [],
        [
          html.span_([], [
            html.h1(css.class([css.color("white")]), [], [
              html.text("Overcomplicated Scroll Pun"),
            ]),
            html.button(
              css.class([
                css.font_size(length.percent(120)),
                css.padding_inline(length.ch(4.0)),
                css.padding_block(length.ch(0.5)),
              ]),
              [evt.on_click(Join)],
              [html.text("LET'S GO")],
            ),
          ]),
        ],
      )
    }
    Game(status) ->
      html.p(
        css.class([
          css.color("white"),
          css.font_weight("bold"),
          css.position("absolute"),
          css.margin(length.px(0)),
          css.text_align("end"),
          css.width(length.percent(100)),
        ]),
        [],
        [
          case status {
            Leader -> html.text("You are in control. Scroll!")
            Line(pos, size) ->
              html.text(
                "#"
                <> int.to_string(pos)
                <> " in line ("
                <> int.to_string(size)
                <> " waiting)",
              )
          },
        ],
      )
  }

  html.div(
    css.class([
      css.position("relative"),
      css.height(length.percent(100)),
      css.border("1.5em solid black"),
      css.box_sizing("border-box"),
    ]),
    [],
    [html.video(vid_class, [attr.id("livekit")], []), game_ui],
  )
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
          css.font_family("-apple-system, helvetica, arial, sans-serif"),
          css.background_color("black"),
        ]),
      )
    })

  let assert Ok(_) =
    lustre.application(init, update, view(_, styles))
    |> lustre.start("body", Nil)
  Nil
}
