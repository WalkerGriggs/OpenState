import React from 'react';
import Heading from './../../components/heading/heading';
import Container from './../../components/container/container';

export default function Features() {
  const component = 'shift-features';

  const data = [
    {
      icon: 'red',
      title: 'Language Agnostic',
      desc: 'Define workflows on your own terms. Callbacks as containers means the sky\'s the limit for libraries and languages.',
    },
    {
      icon: 'purple',
      title: 'Low Code',
      desc: 'Don\'t overthink workflows with serialized, declarative configurations. Creating tasks is as easy as writing a YAML file.',
    },
    {
      icon: 'yellow',
      title: 'Pluggable Runtimes',
      desc: 'OpenState can deploy into any runtime environment, including your existing Kubernetes cluster.'
    },
    {
      icon: 'green',
      title: 'HA and Fault Tolerant',
      desc: 'Built on top of Raft and backed by the Serf gossip protocol, OpenState puts fault tolerance and strong consistency.',
    },
  ];

  const items = data.map((item, index) => {
    const {
      icon,
      title,
      desc,
    } = item;

    return (
      <div className={`${component}__item`} key={index}>
        <div className={`${component}__title ${component}__title--${icon}`}>
          {title}
        </div>
        <div className={`${component}__desc`}>
          {desc}
        </div>
      </div>
    )
  });

  return (
    <div className={component}>
      <Container
        componentClass={component}
        size={'medium'}
      >
        <Heading
          componentClass={component}
          title={'What you need, and nothing more'}
          align={'left'}
        />
        <div className={`${component}__content`}>
          {items}
        </div>
      </Container>
    </div>
  )
}
