import React from 'react';
import useBaseUrl from '@docusaurus/useBaseUrl';
import Heading from './../../components/heading/heading';
import Container from './../../components/container/container';
import Arrow from './../../components/arrow/arrow';

export default function Footer() {

  const component = 'shift-footer';

  const itemsData = [
    {
      icon: useBaseUrl('img/ic-boilerplate.svg'),
      label: 'OpenState<br />Documentation',
      link: '/docs/welcome',
    },
    {
      icon: useBaseUrl('img/ic-boilerplate-plugin.svg'),
      label: 'OpenState<br />Documentation',
      link: '/docs/welcome',
    },
    {
      icon: useBaseUrl('img/ic-frontend-libs.svg'),
      label: 'OpenState<br />Source Code',
      link: 'https://github.com/walkergriggs/openstate',
    },
    {
      icon: useBaseUrl('img/ic-libs.svg'),
      label: 'OpenState<br />Community',
      link: '/community',
    }
  ];

  const items = itemsData.map((item, index) => {
    const {
      icon,
      label,
      link,
    } = item;

    return (
      <div className={`${component}__item`} key={index}>
        <a className={`${component}__link`} href={link} target="_blank" rel="noopener noreferrer">
          <div className={`${component}__icon`}>
            <img src={icon} />
          </div>
          <div className={`${component}__label`} dangerouslySetInnerHTML={{__html: label}}></div>
          <Arrow componentClass={component} />
        </a>
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
          title={'Start exploring'}
        />
        <div className={`${component}__content`}>
          {items}
        </div>
      </Container>
    </div>
  );
}
