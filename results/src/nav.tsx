import './index.css'
import { Component, For, JSXElement, Switch, Match , Show} from 'solid-js'
import { chevronLeft } from "solid-heroicons/solid";
import { Icon } from 'solid-heroicons';
import { A } from '@solidjs/router';
import './index.css'

interface Tab {
  name: string,
  route: string,
}
export const Tab: Component<{ tab: Tab, selected: boolean }> = (props) => {
  return <Switch>
    <Match when={props.selected}> <a class="inline-flex items-center border-b-2 border-indigo-500 px-1 pt-1 text-sm font-medium text-gray-900">{props.tab.name}</a></Match>
    <Match when={!props.selected}> <A href={props.tab.route} class="inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700">{props.tab.name}</A></Match>
  </Switch >
}

export const Nav: Component<{ selected: number, tabs: Tab[] }> = (props) => {
  return <nav class="bg-white shadow">
    <div class="mx-auto px-2 sm:px-6 lg:px-8">
      <div class="relative flex h-16 justify-between">
        <div class="flex flex-1 items-center justify-center sm:items-stretch sm:justify-start">
          <div class="hidden sm:ml-6 sm:flex sm:space-x-8">
            <For each={props.tabs}>{(e, i) => <Tab tab={e} selected={props.selected == i()} />
            }</For>
          </div>
        </div></div></div></nav>
}
export const BackNav: Component<{ children: JSXElement, back: boolean }> = (props) => {
  return <nav class="bg-white shadow px-2">
    <div class="flex flex-row h-12 items-center">
      <Show when={props.back}>
      <Icon onClick={() => history.back()} path={chevronLeft} class='mr-2 h-6 w-6  text-blue-700 hover:text-blue-500' />
      </Show>
      <div class='text-black flex-1  ' >{props.children} </div>

    </div></nav>
}

export function H2(props: { children: JSXElement }) {
    return <h4 class="mx-2 pt-4 pb-2 text-2xl font-bold dark:text-white">{props.children}</h4>
}
export function H3(props: { children: JSXElement }) {
    return <h4 class="pt-4 text-2xl font-bold dark:text-white">{props.children}</h4>
}
export function TagList(props: { each: string[] }) {
    if (!props.each.length) return <div />
    else {
        const [first, ...rest] = props.each
        return <div>
            <For each={rest}>{(e, i) => `, ${e}`}</For>
        </div>
    }
}

function PictureName(props: { path: string }) {
    return <a id={props.path.split('/')[1]} ><h4 class="pt-4 text-2xl font-bold dark:text-white">{props.path.split("/").slice(1).join(": ")}</h4></a>
}