import React from 'react';
import Heading from './../../components/heading/heading';
import Container from './../../components/container/container';

export default function Community() {
    const component = 'shift-community';

    const data = [
        {
            icon: 'green',
            title: 'Slack',
            desc: 'Low-priority, high volume communications. Come and ask questions!',
        },
        {
            icon: 'purple',
            title: 'Email List',
            desc: 'High-priority, low volume annoucements. Includes release information, security notices, and urgent bulletins.'
        },
        {
            icon: 'blue',
            title: 'Bug Tracker',
            desc: 'Use GitHub\'s Issue Tracker. Please only use this for reporting bugs. Do not ask for general help here; use the Slack workspace instead.',
        },
    ]

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
                    title={'Community Resources'}
                    align={'left'}
                />
                <div className={`${component}__content`}>
                    {items}
                </div>
            </Container>
        </div>
    )
}
