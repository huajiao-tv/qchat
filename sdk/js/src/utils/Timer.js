export default class Timer {
    intervalMemory = {
        // 'one': {
        //     id: 12312313,
        //     func: () => {},
        //     time: 0,
        // }
    };

    timeoutMemory = {
        // 'one': {
        //     id: 12312313,
        //     func: () => {},
        //     time: 0,
        // }
    };

    /**
     * 新建 interval 定时器，重名定时器将被覆盖
     * @param name 名称
     * @param func 函数
     * @param time 间隔
     */
    interval(name, func, time) {
        /**
         * 存在则覆盖
         */
        this.removeInterval(name);

        this.intervalMemory[name] = {
            id: setInterval(func, time),
            func,
            time,
        };
        return this.intervalMemory[name].id;
    }

    /**
     * 是否存在指定名称 interval 定时器
     * @param name 名称
     * @return {boolean}
     */
    hasInterval(name) {
        return !!this.intervalMemory[name];
    }

    /**
     * 移除指定名称 interval 定时器
     * @param name 名称
     */
    removeInterval(name) {
        const intervalMemoryElement = this.intervalMemory[name];
        if(intervalMemoryElement) {
            clearInterval(intervalMemoryElement.id);
            delete this.intervalMemory[name];
        }
    }

    /**
     * 移除全部 interval 定时器
     */
    removeAllInterval() {
        for(const key in this.intervalMemory) {
            if(Object.prototype.hasOwnProperty.call(this.intervalMemory, key)) {
                this.removeInterval(key);
            }
        }
    }

    /**
     * 新建 timeout 定时器，重名定时器将被销毁
     *
     * @param name 名称
     * @param func 函数
     * @param time 延迟
     * @param override 是否覆盖上一个同名定时器, 默认覆盖
     */
    timeout(name, func, time, override = true) {
        /**
         * 存在则覆盖
         */
        if(override) {
            this.removeTimeout(name);
        }

        this.timeoutMemory[name] = {
            id: setTimeout(() => {
                // 执行后在列表中移除该 timeout 定时器
                this.removeTimeout(name);
                func();
            }, time),
            func,
            time,
        };
        return this.timeoutMemory[name].id;
    }

    /**
     * 是否存在指定名称 timeout 定时器
     * @param name 名称
     * @return {boolean}
     */
    hasTimeout(name) {
        return !!this.timeoutMemory[name];
    }

    /**
     * 移除指定名称 timeout 定时器
     * @param name 名称
     */
    removeTimeout(name) {
        const timeoutMemoryElement = this.timeoutMemory[name];
        if(timeoutMemoryElement) {
            clearTimeout(timeoutMemoryElement.id);
            delete this.timeoutMemory[name];
        }
    }

    /**
     * 移除全部 timeout 定时器
     */
    removeAllTimeout() {
        for(const key in this.timeoutMemory) {
            if(Object.prototype.hasOwnProperty.call(this.timeoutMemory, key)) {
                this.removeTimeout(key);
            }
        }
    }

    /**
     * 移除全部定时器
     */
    removeAll() {
        this.removeAllInterval();
        this.removeAllTimeout();
    }
}
