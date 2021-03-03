import React from 'react';
import useBaseUrl from '@docusaurus/useBaseUrl';
import Container from './../../components/container/container';

function Foot() {
  const component = 'shift-foot';

  return (
    <div className={component}>
      <Container componentClass={component}>
        <div className={`${component}__columns`}>
          <div className={`${component}__column ${component}__column--left`}>
            {'Made with 🧡 by '}
            <a href="https://walkergriggs.com" target="_blank" rel="noopener noreferrer" className={`${component}__link`}>
              {'Walker Griggs'}
            </a>
          </div>
          <div className={`${component}__column ${component}__column--right`}>
            <a href="https://github.com/walkergriggs" target="_blank" rel="noopener noreferrer" className={`${component}__link`}>
              {'Contact Me'}
            </a>
          </div>
        </div>
      </Container>
    </div>
  );
}

export default Foot;
